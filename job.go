package jobs

import (
	"context"
	"errors"
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
	Create()
	Dispatch(T)
	Dispatches(...T)
	Subscribe()
	handle(T)
	emit()
}

type Job[T any] struct {
	Timeout          time.Duration
	Tries            int
	Delay            time.Duration
	handler          func(T) (any, error)
	handlerSubscribe func(any, error)
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

func (j *Job[T]) Subscribe(handler func(any, error)) *Job[T] {
	j.handlerSubscribe = handler

	return j
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
		j.emit(nil, errors.New("job timeout"))
	}
}

func (j *Job[T]) emit(param any, err error) {
	if j.handlerSubscribe == nil {
		return
	}

	j.handlerSubscribe(param, err)
}
