package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrEmptyDir = errors.New("dir is empty")

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	// Если передан пустой каталог - ошибка
	if dir == "" {
		return nil, ErrEmptyDir
	}

	environment := make(Environment)

	// Читаем файлы в каталоге
	dirItems, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, item := range dirItems {
		// Если не обычный файл или директория - пропускаем
		if !item.Type().IsRegular() {
			continue
		}

		filename := item.Name()

		// Если имя файла содержит '=' - пропускаем
		if strings.Contains(filename, "=") {
			continue
		}

		// Читамем первую строку файла
		value, err := readFirstLine(filepath.Join(dir, filename))
		if err != nil {
			return nil, err
		}

		environment[filename] = EnvValue{
			Value:      value,
			NeedRemove: value == "",
		}
	}

	return environment, nil
}

func readFirstLine(filePath string) (string, error) {
	// Открываем файл с отложенным закрытием
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Читаем первую строку
	if scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, "\x00", "\n")
		line = strings.TrimRight(line, " \t")
		return line, nil
	}

	return "", nil
}
