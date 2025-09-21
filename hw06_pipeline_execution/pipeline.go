package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if in == nil {
		// Если входной канал пустой - возвращаем закрытый входной канал
		result := make(Bi)
		close(result)
		return result
	}

	// Начинаем с входного канала
	out := in
	// Запускаем стадии по цепочке
	for _, stage := range stages {
		// Каждая стадия возвращает свой канал, мы оборачиваем его функции, чтобы использовать done,
		// так как в stage напрямую канал done не передается по сигнатуре
		out = wrapStage(stage(out), done)
	}
	return out
}

func wrapStage(in, done In) Out {
	processOut := make(Bi)

	go func() {
		defer func() {
			// Закрываем выходной канал по завершению работы
			close(processOut)
			// Вычитываем пустым циклом входной канал, чтобы избежать deadlock горутин, запускаемых внутри стадий
			for v := range in {
				_ = v // заглушка для линтера (revive)
			}
		}()

		for {
			select {
			case <-done:
				// Если пришёл сигнал завершения — выходим
				return
			case v, ok := <-in:
				if !ok {
					// Если входной канал закрыт — выходим
					return
				}

				select {
				case <-done:
					// Если во время v, ok := <-in пришёл сигнал done — выходим
					return
				case processOut <- v:
					// Записываем в выходной канал - продолжаем работу
				}
			}
		}
	}()

	return processOut
}
