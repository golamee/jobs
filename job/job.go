package job

import (
	"context"
	"time"
)

func New[T any](handler func(T)) Job[T] {
	return Job[T]{}
}

type JobInterface interface {
	Create()
	Dispatch()
	handle()
	Subscribe()
	emit()
}

type Job[T any] struct {
	Param   T
	Timeout time.Duration
	Tries   int
	handler func(T)
}

func (j *Job[T]) Create(handler func(T)) *Job[T] {
	j.handler = handler

	return j
}

func (j *Job[T]) WithTimeout(timeout time.Duration) *Job[T] {
	j.Timeout = timeout

	return j
}

func (j *Job[T]) Dispatch(param T) *Job[T] {
	j.Param = param

	j.handle()

	return j
}

func (j *Job[T]) handle() *Job[T] {

	var result = make(chan any)

	ctx, err := context.WithTimeout(context.Background(), j.Timeout)

	if err != nil {
		return j
	}

	go j.handler(j.Param)

	select {
	case <-result:
		go j.emit(result)
		return j
	case <-ctx.Done():
		return j
	}
}

func (j *Job[T]) Subscribe(param any) *Job[T] {

	return j
}

func (j *Job[T]) emit(param any) *Job[T] {

	return j
}
