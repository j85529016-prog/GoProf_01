package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("incorrect count of goroutines", func(t *testing.T) {
		err := Run([]Task{}, -1, 10)
		require.ErrorAs(t, err, &ErrWrongCountOfGoroutines)
		err = Run([]Task{}, 0, 10)
		require.ErrorAs(t, err, &ErrWrongCountOfGoroutines)
		err = Run([]Task{}, 1, 10)
		require.NoError(t, err)
	})

	t.Run("incorrect count of errors", func(t *testing.T) {
		err := Run([]Task{}, 10, -1)
		require.ErrorAs(t, err, &ErrWrongCountOfErrors)
		err = Run([]Task{}, 10, 0)
		require.NoError(t, err)
		err = Run([]Task{}, 10, 1)
		require.NoError(t, err)
	})

	t.Run("complex run", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)
		for i := 1; i <= tasksCount; i++ {
			taskID := i
			switch i {
			case 3, 4, 5, 6, 8, 10:
				tasks = append(tasks, func() error {
					return fmt.Errorf("error in func %v", taskID)
				})
			default:
				tasks = append(tasks, func() error {
					return nil
				})
			}
		}
		err := Run(tasks, -1, -1)
		require.Error(t, err)
		err = Run(tasks, 3, 0)
		require.ErrorAs(t, err, &ErrErrorsLimitExceeded)
		err = Run(tasks, 3, 5)
		require.ErrorAs(t, err, &ErrErrorsLimitExceeded)
		err = Run(tasks, 3, 6)
		require.NoError(t, err)
	})

	t.Run("concurrency test", func(t *testing.T) {
		var (
			maxConcurrent int32
			currentActive int32
			mu            sync.Mutex
			taskCount     int
		)

		// Создаем задачи для отслеживания параллельного выполнения
		taskCount = 10
		tasks := make([]Task, taskCount)
		for i := 0; i < taskCount; i++ {
			tasks[i] = func() error {
				// Увеличиваем счетчик активных задач
				active := atomic.AddInt32(&currentActive, 1)
				defer atomic.AddInt32(&currentActive, -1)

				// Обновляем максимум
				mu.Lock()
				if active > maxConcurrent {
					maxConcurrent = active
				}
				mu.Unlock()

				// Какая-то работа
				time.Sleep(50 * time.Millisecond)
				return nil
			}
		}

		go func() {
			_ = Run(tasks, 3, 0)
		}()

		// Проверяем параллельность. Хотя бы 2 горутины работают одновременно
		require.Eventually(
			t,
			func() bool {
				mu.Lock()
				defer mu.Unlock()
				return maxConcurrent >= 2
			},
			500*time.Millisecond,
			50*time.Millisecond,
			"Должно работать минимум 2 горутины одновременно",
		)
	})
}
