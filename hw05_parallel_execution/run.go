package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrErrorsLimitExceeded    = errors.New("errors limit exceeded")
	ErrWrongCountOfGoroutines = errors.New("error: parametr n must be > 0")
)

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrWrongCountOfGoroutines
	}
	if m <= 0 {
		m = 1
	}

	tasksCh := make(chan Task, len(tasks))
	for _, t := range tasks {
		tasksCh <- t
	}
	close(tasksCh)

	var counterErrors int32
	var wg sync.WaitGroup
	wg.Add(n)

	for range n {
		go func() {
			defer wg.Done()
			for task := range tasksCh {
				if int(atomic.LoadInt32(&counterErrors)) >= m {
					return
				}
				err := task()
				if err != nil {
					atomic.AddInt32(&counterErrors, 1)
				}
			}
		}()
	}
	wg.Wait()
	if int(atomic.LoadInt32(&counterErrors)) >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
