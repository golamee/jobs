package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func New[T any](handler ...func(T) (any, error)) *Job[T] {

	j := &Job[T]{}

	if len(handler) > 0 {
		j.handler = handler[0]
	}

	return j
}

type JobInterface[T any] interface {
	Create() JobInterface[T]
	Dispatch(T) JobInterface[T]
	Dispatches(T) JobInterface[T]
	Subscribe() int
	SubscribeOnce() int

	handle(T)
	emit()
}

type subscriber[T any] struct {
	id      string
	handler func(any, error)
	once    bool // ⬅️ Tambahan flag apakah hanya dieksekusi sekali
}

type Job[T any] struct {
	Timeout    time.Duration
	handler    func(T) (any, error)
	subscriber []subscriber[T]
	nextID     int
	mu         sync.Mutex // ⬅️ Mutex untuk proteksi data
}

func (j *Job[T]) Create(handler func(T) (any, error)) *Job[T] {
	j.handler = handler

	return j
}

func (j *Job[T]) WithTimeout(timeout time.Duration) *Job[T] {
	j.Timeout = timeout

	return j
}

func (j *Job[T]) Dispatch(param T) *Job[T] {

	j.handle(param)

	return j
}

func (j *Job[T]) Dispatches(params ...T) *Job[T] {

	for _, param := range params {
		j.handle(param)
	}

	return j
}

func (j *Job[T]) Subscribe(handler func(any, error)) string {
	j.mu.Lock()
	defer j.mu.Unlock()

	id := fmt.Sprintf("sub-%d", j.nextID)
	j.nextID++

	j.subscriber = append(j.subscriber, subscriber[T]{
		id:      id,
		handler: handler,
		once:    false,
	})
	return id
}

func (j *Job[T]) Unsubscribe(id string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for i, sub := range j.subscriber {
		if sub.id == id {
			j.subscriber = append(j.subscriber[:i], j.subscriber[i+1:]...)
			break
		}
	}
}

func (j *Job[T]) SubscribeOnce(handler func(any, error)) string {

	j.mu.Lock()

	defer j.mu.Unlock()

	id := fmt.Sprintf("sub-%d", j.nextID)
	j.nextID++

	j.subscriber = append(j.subscriber, subscriber[T]{
		id:      id,
		handler: handler,
		once:    true,
	})

	return id
}

func (j *Job[T]) handle(param T) {
	resultChan := make(chan any, 1)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), j.Timeout)
	defer cancel()

	go func() {

		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.New("panic occurred in handler")
			}
		}()

		res, err := j.handler(param)

		if err != nil {
			errChan <- err
			return
		}
		resultChan <- res
	}()

	select {
	case res := <-resultChan:
		j.emit(res, nil)
	case err := <-errChan:
		j.emit(nil, err)
	case <-ctx.Done():
		j.emit(nil, errors.New("Job timeout"))
	}
}

func (j *Job[T]) emit(param any, err error) {

	j.mu.Lock()

	subs := make([]subscriber[T], len(j.subscriber))

	copy(subs, j.subscriber) // Copy dulu untuk dibaca di luar lock

	j.mu.Unlock()

	remaining := make([]subscriber[T], 0, len(subs))

	for _, sub := range subs {
		sub.handler(param, err)
		if !sub.once {
			remaining = append(remaining, sub)
		}
	}

	j.mu.Lock()
	j.subscriber = remaining
	j.mu.Unlock()
}
