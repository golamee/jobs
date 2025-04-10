package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func NewJob[T any](handler ...func(T) (any, error)) *Job[T] {
	j := &Job[T]{
		// Timeout: 10 * time.Second,
		// Delay: 3 * time.Second,
	}

	if len(handler) > 0 {
		j.handler = handler[0]
	}

	return j
}

type JobInterface[T any] interface {
	Create() *Job[T]
	Dispatch(T) *Job[T]
	Dispatches(T) *Job[T]

	WithTimeout(time.Duration) *Job[T]
	WithDelay(time.Duration) *Job[T]

	Subscribe(func(any), func(error)) int
	SubscribeOnce(func(any), func(error)) int

	handle(T)
	emit()
}

type subscriber[T any] struct {
	id            string
	handler       func(any)
	handlerOnFail func(error)
	once          bool // ⬅️ Tambahan flag apakah hanya dieksekusi sekali
}

type Job[T any] struct {
	Timeout time.Duration
	Delay   time.Duration

	handler    func(T) (any, error)
	subscriber []subscriber[T]

	nextID int
	mu     sync.Mutex // ⬅️ Mutex untuk proteksi data
}

func (j *Job[T]) Create(handler func(T) (any, error)) *Job[T] {
	j.handler = handler
	return j
}

func (j *Job[T]) WithTimeout(timeout time.Duration) *Job[T] {
	j.Timeout = timeout
	return j
}

func (j *Job[T]) WithDelay(delay time.Duration) *Job[T] {
	j.Delay = delay
	return j
}

func (j *Job[T]) Dispatch(param T) *Job[T] {
	go j.handle(param)
	return j
}

func (j *Job[T]) Dispatches(params ...T) *Job[T] {
	for _, param := range params {
		go j.handle(param)
	}
	return j
}

func (j *Job[T]) handle(param T) {
	resultChan := make(chan any, 1)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), j.Timeout) // Implement timeout
	defer cancel()

	var run func() = func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.New("panic occurred in job's handler")
			}
		}()

		time.Sleep(j.Delay) // Implement Delay

		res, err := j.handler(param)

		if err != nil {
			errChan <- err
			return
		}
		resultChan <- res

	}

	go run()

	select {
	case res := <-resultChan:
		j.emit(res, nil)
	case err := <-errChan:
		j.emit(nil, err)
	case <-ctx.Done():
		j.emit(nil, errors.New("Job timeout"))
	}
}

func (j *Job[T]) Subscribe(onSuccess func(any), onFail func(error)) string {

	j.mu.Lock()
	defer j.mu.Unlock()

	id := fmt.Sprintf("sub-%d", j.nextID)
	j.nextID++

	newSubscriber := subscriber[T]{
		id:            id,
		handler:       onSuccess,
		handlerOnFail: onFail,
		once:          false,
	}

	j.subscriber = append(j.subscriber, newSubscriber)

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

func (j *Job[T]) SubscribeOnce(onSuccess func(any), onFail func(error)) string {
	j.mu.Lock()
	defer j.mu.Unlock()

	id := fmt.Sprintf("sub-%d", j.nextID)
	j.nextID++

	newSubscriber := subscriber[T]{
		id:            id,
		handler:       onSuccess,
		handlerOnFail: onFail,
		once:          true,
	}

	j.subscriber = append(j.subscriber, newSubscriber)

	return id
}

func (j *Job[T]) emit(param any, err error) {
	if len(j.subscriber) == 0 {
		return
	}

	j.mu.Lock()

	subs := make([]subscriber[T], len(j.subscriber))

	copy(subs, j.subscriber) // Copy dulu untuk dibaca di luar lock

	j.mu.Unlock()

	remaining := make([]subscriber[T], 0, len(subs))

	for _, sub := range subs {

		if err == nil {
			sub.handler(param)
		} else {
			sub.handlerOnFail(err)
		}

		if !sub.once {
			remaining = append(remaining, sub)
		}
	}

	j.mu.Lock()
	j.subscriber = remaining
	j.mu.Unlock()
}
