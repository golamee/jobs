package jobs

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Man struct {
	Name string
	Age  int
}

func TestJob(t *testing.T) {
	t.Run("CreateOnly", func(t *testing.T) {

		j := NewJob[Man]()

		j.Create(func(value Man) (any, error) {
			return fmt.Sprintf("Name: %s, Age: %d", value.Name, value.Age), nil
		})

		assert.NotNil(t, j, "Job should not be nil after Create")

	})

	t.Run("DispatchFailed", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		expectedName := "John Doe"
		expectedAge := 30
		expectedErr := errors.New("simulated error")

		job := NewJob[Man]().
			Create(func(value Man) (any, error) {

				defer wg.Done()

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return nil, expectedErr
			}).
			WithTimeout(2 * time.Second)

		job.Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

	t.Run("DispatchSuccess", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		expectedName := "John Doe"
		expectedAge := 30

		job := NewJob[Man]().
			Create(func(value Man) (any, error) {

				defer wg.Done()

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return value.Name, nil
			}).
			WithTimeout(2 * time.Second)

		job.Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

}

func TestJobWithSubscribe(t *testing.T) {

	t.Run("DispatchFailed:Subscribe", func(t *testing.T) {
		var wg sync.WaitGroup

		expectedName := "John Doe"
		expectedAge := 30
		expectedErr := errors.New("simulated error")

		job := NewJob[Man]().
			Create(func(value Man) (any, error) {

				wg.Add(1)

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return nil, expectedErr
			}).
			WithTimeout(2 * time.Second)

		job.Subscribe(func(result any) {

			assert.True(t, false, "Subscribe should not be called")
		}, func(err error) {

			defer wg.Done()

			assert.NotNil(t, err, "Err should not be nil")
			assert.EqualError(t, err, expectedErr.Error(), "Err should match the expected value")
		})

		job.Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

	t.Run("DispatchSuccess:Subscribe", func(t *testing.T) {
		var wg sync.WaitGroup

		expectedName := "John Doe"
		expectedAge := 30

		job := NewJob[Man]().
			Create(func(value Man) (any, error) {

				wg.Add(1)

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return value.Name, nil
			}).
			WithTimeout(2 * time.Second)

		job.Subscribe(func(result any) {

			defer wg.Done()

			name, ok := result.(string)

			assert.True(t, ok, "Result should be a string")
			assert.Equal(t, expectedName, name, "Name should match the expected value")

		}, func(err error) {

			assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))
		})

		job.Dispatch(Man{Name: expectedName, Age: expectedAge})
		job.Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

	t.Run("DispatchesSuccess:Subscribe", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(6)

		expectedName := "John Doe"
		expectedAge := 30

		job := NewJob[Man]().
			Create(func(value Man) (any, error) {

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return value.Name, nil
			}).
			WithTimeout(2 * time.Second)

		job.Subscribe(func(result any) {

			defer wg.Done()

			name, ok := result.(string)

			assert.True(t, ok, "Result should be a string")
			assert.Equal(t, expectedName, name, "Name should match the expected value")

		}, func(err error) {

			assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))
		})

		job.Dispatches(Man{Name: expectedName, Age: expectedAge}, Man{Name: expectedName, Age: expectedAge}, Man{Name: expectedName, Age: expectedAge})
		job.Dispatches(Man{Name: expectedName, Age: expectedAge}, Man{Name: expectedName, Age: expectedAge}, Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

	t.Run("DispatchSuccess:Subscribes", func(t *testing.T) {
		var wg sync.WaitGroup

		expectedName := "John Doe"
		expectedAge := 30

		job := NewJob[Man]().
			Create(func(value Man) (any, error) {

				wg.Add(3)

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return value.Name, nil
			}).
			WithTimeout(2 * time.Second)

		job.Subscribe(func(result any) {

			defer wg.Done()

			name, ok := result.(string)

			assert.True(t, ok, "Result should be a string")
			assert.Equal(t, expectedName, name, "Name should match the expected value")

		}, func(err error) {

			assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))
		})

		job.Subscribe(func(result any) {

			defer wg.Done()

			name, ok := result.(string)

			assert.True(t, ok, "Result should be a string")
			assert.Equal(t, expectedName, name, "Name should match the expected value")

		}, func(err error) {

			assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))
		})

		job.Subscribe(func(result any) {

			defer wg.Done()

			name, ok := result.(string)

			assert.True(t, ok, "Result should be a string")
			assert.Equal(t, expectedName, name, "Name should match the expected value")

		}, func(err error) {

			assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))
		})

		job.Dispatch(Man{Name: expectedName, Age: expectedAge})
		job.Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

	t.Run("DispatchSuccess:SubscribeOnce", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(4)

		expectedName := "John Doe"
		expectedAge := 30

		counter := 1

		job := NewJob[Man]().
			Create(func(value Man) (any, error) {

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return value.Name, nil
			}).
			WithTimeout(2 * time.Second)

		job.Subscribe(func(result any) {

			defer wg.Done()

			name, ok := result.(string)

			assert.True(t, ok, "Result should be a string")
			assert.Equal(t, expectedName, name, "Name should match the expected value")

		}, func(err error) {

			assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))
		})

		job.SubscribeOnce(func(result any) {

			defer wg.Done()

			wg.Add(1)

			name, ok := result.(string)

			assert.True(t, ok, "Result should be a string")
			assert.Equal(t, expectedName, name, "Name should match the expected value")

			assert.Greater(t, counter, 0, "Counter should be greater than 0")
			counter--
		}, func(err error) {

			assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))
		})

		job.Dispatch(Man{Name: expectedName, Age: expectedAge})
		job.Dispatch(Man{Name: expectedName, Age: expectedAge})
		job.Dispatch(Man{Name: expectedName, Age: expectedAge})
		job.Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})
}
