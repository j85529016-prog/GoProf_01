package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrErrorsLimitExceeded    = errors.New("errors limit exceeded")
	ErrWrongCountOfGoroutines = errors.New("error: parametr n must be > 0")
	ErrWrongCountOfErrors     = errors.New("error: parametr m must be >= 0")
)

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	// Проверяем пограничные случаи для немедленного завершения функции
	if n <= 0 {
		return ErrWrongCountOfGoroutines
	}
	if m < 0 {
		return ErrWrongCountOfErrors
	}

	// Берем в переменную общее число тасок
	countTasksTotal := len(tasks)

	// Cоздаем канал тасок, заполняем и закрываем его (чтобы при чтении из канала не возник deadlock)
	tasksCh := make(chan Task, countTasksTotal)
	for _, t := range tasks {
		tasksCh <- t
	}
	close(tasksCh)

	// Определяем кол-во групп тасок
	var countGroups int
	if countTasksTotal-countTasksTotal/n == 0 {
		countGroups = countTasksTotal / n
	} else {
		countGroups = countTasksTotal/n + 1
	}

	// Cоздаем atomic счетчик ошибок
	var counterErrors int32

	// Итерируемся по группам тасок, запуская их в горутинах
	for i := range countGroups {
		// Определяем число тасок в текущей группе
		countTasksInCurrentgroup := min(countTasksTotal-i*n, n)

		var wg sync.WaitGroup

		for range countTasksInCurrentgroup {
			// Оборачиваем таску в доп. логику
			taskExt := func() {
				defer wg.Done()
				// Считываем функцию из канала и запускаем ее
				taskToRun := <-tasksCh
				err := taskToRun()
				// Если таска вернула ошибку - увеличиваем атомарный счетчик ошибок
				if err != nil {
					atomic.AddInt32(&counterErrors, 1)
				}
			}

			// Если не превысили лимит по ошибкам - запускаем горутину таски, иначе прерываем цикл
			if int(atomic.LoadInt32(&counterErrors)) <= m {
				wg.Add(1)
				go taskExt()
			} else {
				return ErrErrorsLimitExceeded
			}
		}
		wg.Wait()
	}

	return nil
}
