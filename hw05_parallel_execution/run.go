package hw05parallelexecution

import (
	"errors"
	"sync"
)

var (
	ErrErrorsLimitExceeded    = errors.New("errors limit exceeded")
	ErrWrongCountOfGoroutines = errors.New("error: parametr n must be > 0")
	ErrWrongCountOfErrors     = errors.New("error: parametr m must be >= 0")
)

type counterErrors struct {
	counter int
	mutex   sync.Mutex
}
type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	// n - кол-во одновременно выполняющихся горутин
	// m - максимальное число ошибок

	// Обрабатываем пограничные случаи
	if n <= 0 {
		return ErrWrongCountOfGoroutines
	}
	if m < 0 {
		return ErrWrongCountOfErrors
	}

	// Создаем структуру с мьютексом для записи кол-ва ошибок
	counterErrors := counterErrors{}

	// Слайс с группами количества горутин одновременного выполнения для последовательного запуска в количестве <=n
	groupsCountTasks := getIntParts(len(tasks), n)

	// Итерируемся по группам кол-ва горутин для запуска
	counter := -1
	for _, v := range groupsCountTasks {
		var wg sync.WaitGroup
		wg.Add(v)
		// Итерируемся по таскам
		for j := 1; j <= v; j++ {
			// Номер таски по проядку в первоначальном слайсе
			counter++

			// Оборачиваем таску в доп. логику
			task := func(numberTask int, wg *sync.WaitGroup) {
				err := tasks[numberTask]()
				// Если таска вернула ошибку - изменяем счетчик ошибок
				if err != nil {
					counterErrorsAdd(&counterErrors)
				}
				wg.Done()
			}

			// Если не достигли лимита по ошибкам - запускаем таску, иначе прерываем цикл
			if !counterErrorsHaveLimit(&counterErrors, m) {
				go task(counter, &wg)
			} else {
				break
			}
		}
		wg.Wait()
		if counterErrorsHaveLimit(&counterErrors, m) {
			return ErrErrorsLimitExceeded
		}
	}

	return nil
}

func getIntParts(countAll, countInPart int) []int {
	result := make([]int, 0)
	intParts := countAll/countInPart + 1
	for i := 1; i <= intParts; i++ {
		if i == intParts {
			result = append(result, countAll-(intParts-1)*countInPart)
			break
		}
		result = append(result, countInPart)
	}
	return result
}

// Потокобезобасно проверяет, достиг ли счетчик ошибок лимита ошибок по их количеству.
func counterErrorsHaveLimit(mx *counterErrors, limit int) bool {
	mx.mutex.Lock()
	defer mx.mutex.Unlock()

	return mx.counter > limit
}

// Потокобезобасно увеличивает счетчик ошибок.
func counterErrorsAdd(mx *counterErrors) {
	mx.mutex.Lock()
	defer mx.mutex.Unlock()

	mx.counter++
}
