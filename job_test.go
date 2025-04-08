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

		j := New[Man]()

		j.Create(func(value Man) (any, error) {
			return fmt.Sprintf("Name: %s, Age: %d", value.Name, value.Age), nil
		})

		assert.NotNil(t, j, "Job should not be nil after Create")

	})

	t.Run("DispatchSuccess", func(t *testing.T) {
		var wg sync.WaitGroup

		expectedName := "John Doe"
		expectedAge := 30

		New[Man]().
			Create(func(value Man) (any, error) {

				wg.Add(1)

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return value.Name, nil
			}).
			WithTimeout(2 * time.Second).
			Subscribe(func(result any, err error) {

				defer wg.Done()

				assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))

				name, ok := result.(string)

				assert.True(t, ok, "Result should be a string")
				assert.Equal(t, expectedName, name, "Name should match the expected value")

			}).
			Dispatch(Man{Name: expectedName, Age: expectedAge}).
			Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

	t.Run("DispatchFailed", func(t *testing.T) {
		var wg sync.WaitGroup

		expectedName := "John Doe"
		expectedAge := 30
		expectedErr := errors.New("simulated error")

		New[Man]().
			Create(func(value Man) (any, error) {

				wg.Add(1)

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return nil, expectedErr
			}).
			WithTimeout(2 * time.Second).
			Subscribe(func(result any, err error) {

				defer wg.Done()

				assert.NotNil(t, err, "Err should not be nil")
				assert.EqualError(t, err, expectedErr.Error(), "Err should match the expected value")

			}).
			Dispatch(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})

	t.Run("DispatchesSuccess", func(t *testing.T) {
		var wg sync.WaitGroup

		expectedName := "John Doe"
		expectedAge := 30

		New[Man]().
			Create(func(value Man) (any, error) {

				wg.Add(1)

				assert.Equal(t, expectedName, value.Name, "Name should match the expected value")
				assert.Equal(t, expectedAge, value.Age, "Age should match the expected value")

				return value.Name, nil
			}).
			WithTimeout(2 * time.Second).
			Subscribe(func(result any, err error) {

				defer wg.Done()

				assert.Nil(t, err, fmt.Sprintf("Err should be nil. Unexpected error: %v", err))

				name, ok := result.(string)

				assert.True(t, ok, "Result should be a string")
				assert.Equal(t, expectedName, name, "Name should match the expected value")

			}).
			Dispatches(Man{Name: expectedName, Age: expectedAge})

		wg.Wait()
	})
}
