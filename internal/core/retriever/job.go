package retriever

import (
	"fmt"
	"sync"
)

// Result wraps a result with potential error
type Result[T any] struct {
	Data T
	Err  error
}

// ExecuteJobs runs jobs concurrently and collects results
// TJob is the input type for each job, TResult is the output type
func ExecuteJobs[TJob any, TResult any](
	jobs []TJob,
	workers int,
	worker func(TJob) ([]TResult, error),
) []Result[[]TResult] {
	results := make([]Result[[]TResult], len(jobs))
	jobCh := make(chan int)
	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {
			for i := range jobCh {
				func() {
					defer func() {
						if r := recover(); r != nil {
							results[i].Err = fmt.Errorf("panic: %v", r)
						}
					}()

					data, err := worker(jobs[i])
					results[i] = Result[[]TResult]{Data: data, Err: err}
				}()
			}
		})
	}

	for i := range jobs {
		jobCh <- i
	}
	close(jobCh)

	wg.Wait()
	return results
}

// AggregateResults collects all successful results and returns first error if any
func AggregateResults[T any](results []Result[[]T]) ([]T, error) {
	var aggregated []T
	for _, res := range results {
		if res.Err != nil {
			return nil, res.Err
		}
		aggregated = append(aggregated, res.Data...)
	}
	return aggregated, nil
}
