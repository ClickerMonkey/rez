package db

import (
	"context"
	"sync"
)

// A type which runs a query on a given connection.
type Runnable interface {
	Run(ctx context.Context, able Queryable) error
}

// A runner is a function which processes many runnable interface
type Runner func(many []Runnable, ctx context.Context, able Queryable) error

// Runs a list of queries sequentially
func RunSequential(many []Runnable, ctx context.Context, able Queryable) error {
	for _, run := range many {
		err := run.Run(ctx, able)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ Runner = RunSequential

// Runs a list of queries at the same time. If any of them errors it will
// try to stop all other queries and return as soon as possible.
func RunConcurrently(many []Runnable, ctx context.Context, able Queryable) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	lastError := make(chan error)
	group := sync.WaitGroup{}

	for i := range many {
		group.Add(1)
		go func(run Runnable) {
			defer group.Done()
			err := run.Run(cancelCtx, able)
			if err != nil {
				cancel()
				lastError <- err
				close(lastError)
			}
		}(many[i])
	}
	group.Wait()
	cancel()

	select {
	case err := <-lastError:
		return err
	default:
	}
	return nil
}

var _ Runner = RunConcurrently

// Creates a pooled runner
func NewPooledRunner(poolSize int) Runner {
	return func(many []Runnable, ctx context.Context, able Queryable) error {
		return RunPooled(many, ctx, able, poolSize)
	}
}

// Runs a list of queries at the same time. If any of them errors it will
// try to stop all other queries and return as soon as possible.
func RunPooled(many []Runnable, ctx context.Context, able Queryable, poolSize int) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	lastError := make(chan error)
	group := sync.WaitGroup{}

	jobs := make(chan struct{}, poolSize)
	for i := 0; i < poolSize; i++ {
		jobs <- struct{}{}
	}

	for i := range many {
		group.Add(1)
		go func(run Runnable) {
			defer group.Done()
			<-jobs
			err := run.Run(cancelCtx, able)
			if err != nil {
				cancel()
				lastError <- err
				close(lastError)
			} else {
				jobs <- struct{}{}
			}
		}(many[i])
	}
	group.Wait()
	close(jobs)
	cancel()

	select {
	case err := <-lastError:
		return err
	default:
	}

	return nil
}

var _ Runner = NewPooledRunner(10)
