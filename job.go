package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func NewJob[T any, S any](handler ...func(T) (S, error)) *Job[T, S] {
	j := &Job[T, S]{
		// Timeout: 10 * time.Second,
		// Delay: 3 * time.Second,
	}

	if len(handler) > 0 {
		j.handler = handler[0]
	}

	return j
}

type JobInterface[T any, S any] interface {
	Create() *Job[T, S]
	Dispatch(T) *Job[T, S]
	Dispatches(T) *Job[T, S]

	WithTimeout(time.Duration) *Job[T, S]
	WithDelay(time.Duration) *Job[T, S]

	Subscribe(func(any), func(error)) int
	SubscribeOnce(func(any), func(error)) int

	handle(T)
	emitSuccess(T)
	emitFailed(error)
}

type subscriber[S any] struct {
	id            string
	handler       func(S)
	handlerOnFail func(error)
	once          bool // ⬅️ Tambahan flag apakah hanya dieksekusi sekali
}

type Job[T any, S any] struct {
	Timeout time.Duration
	Delay   time.Duration

	handler    func(T) (S, error)
	subscriber []subscriber[S]

	nextID int
	mu     sync.Mutex // ⬅️ Mutex untuk proteksi data
}

func (j *Job[T, S]) Create(handler func(T) (S, error)) *Job[T, S] {
	j.handler = handler
	return j
}

func (j *Job[T, S]) WithTimeout(timeout time.Duration) *Job[T, S] {
	j.Timeout = timeout
	return j
}

func (j *Job[T, S]) WithDelay(delay time.Duration) *Job[T, S] {
	j.Delay = delay
	return j
}

func (j *Job[T, S]) Dispatch(param T) *Job[T, S] {
	time.Sleep(j.Delay) // Implement Delay

	go j.handle(param)
	return j
}

func (j *Job[T, S]) Dispatches(params ...T) *Job[T, S] {
	time.Sleep(j.Delay) // Implement Delay

	for _, param := range params {
		go j.handle(param)
	}
	return j
}

func (j *Job[T, S]) handle(param T) {
	resultChan := make(chan S, 1)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), j.Timeout) // Implement timeout
	defer cancel()

	var run func() = func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- errors.New("panic occurred in job's handler")
			}
		}()

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
		j.emitSuccess(res)
	case err := <-errChan:
		j.emitFailed(err)
	case <-ctx.Done():
		j.emitFailed(errors.New("Job timeout"))
	}
}

func (j *Job[T, S]) Subscribe(onSuccess func(S), onFail func(error)) string {

	j.mu.Lock()
	defer j.mu.Unlock()

	id := fmt.Sprintf("sub-%d", j.nextID)
	j.nextID++

	newSubscriber := subscriber[S]{
		id:            id,
		handler:       onSuccess,
		handlerOnFail: onFail,
		once:          false,
	}

	j.subscriber = append(j.subscriber, newSubscriber)

	return id
}

func (j *Job[T, S]) Unsubscribe(id string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for i, sub := range j.subscriber {
		if sub.id == id {
			j.subscriber = append(j.subscriber[:i], j.subscriber[i+1:]...)
			break
		}
	}
}

func (j *Job[T, S]) SubscribeOnce(onSuccess func(S), onFail func(error)) string {
	j.mu.Lock()
	defer j.mu.Unlock()

	id := fmt.Sprintf("sub-%d", j.nextID)
	j.nextID++

	newSubscriber := subscriber[S]{
		id:            id,
		handler:       onSuccess,
		handlerOnFail: onFail,
		once:          true,
	}

	j.subscriber = append(j.subscriber, newSubscriber)

	return id
}

func (j *Job[T, S]) emitSuccess(param S) {
	if len(j.subscriber) == 0 {
		return
	}

	j.mu.Lock()
	subs := make([]subscriber[S], len(j.subscriber))
	copy(subs, j.subscriber) // Copy dulu untuk dibaca di luar lock
	j.mu.Unlock()

	remaining := make([]subscriber[S], 0, len(subs))

	j.mu.Lock()
	for _, sub := range subs {

		sub.handler(param)

		if !sub.once {
			remaining = append(remaining, sub)
		}
	}
	j.mu.Unlock()

	j.mu.Lock()
	j.subscriber = remaining
	j.mu.Unlock()
}

func (j *Job[T, S]) emitFailed(err error) {
	if len(j.subscriber) == 0 {
		return
	}

	j.mu.Lock()
	subs := make([]subscriber[S], len(j.subscriber))
	copy(subs, j.subscriber) // Copy dulu untuk dibaca di luar lock
	j.mu.Unlock()

	remaining := make([]subscriber[S], 0, len(subs))

	j.mu.Lock()
	for _, sub := range subs {

		sub.handlerOnFail(err)

		if !sub.once {
			remaining = append(remaining, sub)
		}
	}
	j.mu.Unlock()

	j.mu.Lock()
	j.subscriber = remaining
	j.mu.Unlock()
}
