// parallel.go — ordered, bounded worker orchestration for market workloads.
package market

import (
	"fmt"
	"sync"
)

func resolveWorkerLimit(cliWorkers, globalWorkers, perHostWorkers, itemCount int) (int, error) {
	if cliWorkers < 0 {
		return 0, fmt.Errorf("--concurrency must be >= 0")
	}
	if itemCount <= 0 {
		return 0, nil
	}
	if globalWorkers < 1 {
		return 0, fmt.Errorf("concurrency.global_workers must be >= 1")
	}
	if perHostWorkers < 1 {
		return 0, fmt.Errorf("concurrency.per_host_workers must be >= 1")
	}

	workers := globalWorkers
	if cliWorkers > 0 {
		workers = cliWorkers
	}
	if workers > perHostWorkers {
		workers = perHostWorkers
	}
	if workers > itemCount {
		workers = itemCount
	}
	return workers, nil
}

func mapBounded[Input, Output any](items []Input, workers int, transform func(Input) Output) []Output {
	results := make([]Output, len(items))
	if len(items) == 0 {
		return results
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(items) {
		workers = len(items)
	}

	jobs := make(chan int)
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			for index := range jobs {
				results[index] = transform(items[index])
			}
		}()
	}

	for index := range items {
		jobs <- index
	}
	close(jobs)
	group.Wait()
	return results
}
