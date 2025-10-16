package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	// Получаем текущее окружение
	currentEnv := os.Environ()
	envMap := make(map[string]string)

	// Заполняем мапу текущим окружением
	for _, kv := range currentEnv {
		if i := strings.IndexByte(kv, '='); i >= 0 {
			envMap[kv[:i]] = kv[i+1:]
		}
	}

	// Применяем изменения из env
	for name, envVal := range env {
		if envVal.NeedRemove {
			// Удаляем переменную
			delete(envMap, name)
		} else {
			// Устанавливаем новое значение
			envMap[name] = envVal.Value
		}
	}

	// Собираем обратно в []string
	newEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		newEnv = append(newEnv, k+"="+v)
	}

	// #nosec G204
	command := exec.Command(cmd[0], cmd[1:]...)
	command.Env = newEnv
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		var exitError *exec.ExitError
		if ok := errors.Is(err, exitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}
