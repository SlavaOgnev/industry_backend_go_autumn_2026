package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(in) == 0 {
		return []Result[R]{}, nil
	}
	results := make([]Result[R], len(in))
	countWorkers := min(workers, len(in))

	var jobs = make(chan int)

	wg := sync.WaitGroup{}
	wg.Add(countWorkers)

	for range countWorkers {
		go func() {
			defer wg.Done()

			for x := range jobs {
				if ctx.Err() == nil {
					r, err := fn(ctx, in[x])
					if err != nil {
						results[x].Err = err
					} else {
						results[x] = Result[R]{r, err}
					}
				}
			}
		}()
	}

	for i := range in {
		select {
		case <-ctx.Done():

		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return results, nil
}
