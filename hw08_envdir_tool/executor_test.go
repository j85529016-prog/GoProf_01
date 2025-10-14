package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	// Сохраняем исходное состояние используемых в тесте переменных окружения
	origHELLO, helloExists := os.LookupEnv("HELLO")
	origFOO, fooExists := os.LookupEnv("FOO")
	origUNSET, unsetExists := os.LookupEnv("UNSET")
	origADDED, addedExists := os.LookupEnv("ADDED")
	origEMPTY, emptyExists := os.LookupEnv("EMPTY")

	// Устанавливаем тестовое окружение (как в test.sh)
	os.Setenv("HELLO", "SHOULD_REPLACE")
	os.Setenv("FOO", "SHOULD_REPLACE")
	os.Setenv("UNSET", "SHOULD_REMOVE")
	os.Setenv("ADDED", "from original env")
	os.Setenv("EMPTY", "SHOULD_BE_EMPTY")

	// Восстанавливаем исходное состояние после теста
	defer func() {
		if helloExists {
			os.Setenv("HELLO", origHELLO)
		} else {
			os.Unsetenv("HELLO")
		}
		if fooExists {
			os.Setenv("FOO", origFOO)
		} else {
			os.Unsetenv("FOO")
		}
		if unsetExists {
			os.Setenv("UNSET", origUNSET)
		} else {
			os.Unsetenv("UNSET")
		}
		if addedExists {
			os.Setenv("ADDED", origADDED)
		} else {
			os.Unsetenv("ADDED")
		}
		if emptyExists {
			os.Setenv("EMPTY", origEMPTY)
		} else {
			os.Unsetenv("EMPTY")
		}
	}()

	// Читаем Environment из testdata
	env, err := ReadDir("testdata/env")
	require.NoError(t, err, "ReadDir must not fail")

	// Применяем логику RunCmd: строим новое окружение
	currentEnv := os.Environ()
	envMap := make(map[string]string)
	for _, kv := range currentEnv {
		if i := strings.IndexByte(kv, '='); i >= 0 {
			envMap[kv[:i]] = kv[i+1:]
		}
	}

	for name, ev := range env {
		if ev.NeedRemove {
			delete(envMap, name)
		} else {
			envMap[name] = ev.Value
		}
	}

	// Проверяем ожидаемые переменные
	require.Equal(t, "\"hello\"", envMap["HELLO"], "HELLO should be overwritten from file")
	require.Equal(t, "bar", envMap["BAR"], "BAR should be added from file")
	require.Equal(t, "   foo\nwith new line", envMap["FOO"], "FOO should contain newline from \\x00")
	require.Equal(t, "from original env", envMap["ADDED"], "ADDED should remain from original env")
	require.Equal(t, "", envMap["EMPTY"], "EMPTY should be set to empty string")

	// Проверяем, что UNSET удалена
	_, exists := envMap["UNSET"]
	require.False(t, exists, "UNSET should be removed because file is empty")
}
