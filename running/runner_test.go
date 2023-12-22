package running

import (
	"context"
	"errors"
	"testing"
	"time"
)

type RunnerFunc func(ctx context.Context) error

func (f RunnerFunc) Run(ctx context.Context) error {
	return f(ctx)
}
func (f RunnerFunc) Shutdown() {}

func TestRunner(t *testing.T) {
	group := NewGroup(
		RunnerFunc(func(ctx context.Context) error {
			time.Sleep(time.Second)
			return nil
		}),
		RunnerFunc(func(ctx context.Context) error {
			time.Sleep(2 * time.Second)
			return nil
		}),
	)
	group.Run()
	for {
		select {
		case err := <-group.Error():
			t.Error(err)
		case <-group.Done():
			t.Log("done")
			return
		}
	}
}

func TestRunnerError(t *testing.T) {
	var expectedError = errors.New("some error")
	group := NewGroup(
		RunnerFunc(func(ctx context.Context) error {
			time.Sleep(time.Second)
			return expectedError
		}),
		RunnerFunc(func(ctx context.Context) error {
			time.Sleep(3 * time.Second)
			return expectedError
		}),
		RunnerFunc(func(ctx context.Context) error {
			time.Sleep(5 * time.Second)
			return expectedError
		}),
	)
	group.Run()
	for {
		select {
		case err := <-group.Error():
			t.Log(err)
		case <-group.Done():
			t.Log("done")
			return
		}
	}
}
