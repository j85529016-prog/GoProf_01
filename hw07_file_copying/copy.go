package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrOffPaths              = errors.New("from and to paths must be specified")
	ErrNegativeLimitOrOffset = errors.New("limit and offset must be positive or zero")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	// Окрываем исходный файл с проверками и отложенным закрытием
	sourceFile, err := checkAndOpenFile(fromPath, toPath, offset, limit)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Создаем целевой файл
	outputFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Перемещаем каретку в нужную позицию
	_, err = sourceFile.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	// Определяем сколько байт нужно скопировать
	var bytesToCopy int64
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	sourceFileSize := sourceFileStat.Size()
	if limit == 0 {
		bytesToCopy = sourceFileSize - offset
	} else {
		bytesToCopy = min(limit, sourceFileSize-offset)
	}

	// Копируем данные
	if bytesToCopy > 0 {
		bar := pb.Full.Start64(bytesToCopy)
		bar.Set(pb.Bytes, true)
		barWriter := bar.NewProxyWriter(outputFile)

		_, err = io.CopyN(barWriter, sourceFile, bytesToCopy)
		bar.Finish()

		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("copy error: %w", err)
		}
	}

	return nil
}

func checkAndOpenFile(fromPath, toPath string, offset, limit int64) (*os.File, error) {
	// Пустые пути - ошибка
	if fromPath == "" || toPath == "" {
		return nil, ErrOffPaths
	}

	// Отрицательные limit и offset - ошибка
	if offset < 0 || limit < 0 {
		return nil, ErrNegativeLimitOrOffset
	}

	// Ошибка при открытии файла
	file, err := os.OpenFile(fromPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	// Ошибка получения информации о файле
	fileStat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error checking file: %w", err)
	}

	// Неподдерживаемый файл - ошибка
	if !fileStat.Mode().IsRegular() {
		file.Close()
		return nil, ErrUnsupportedFile
	}
	if fileStat.Size() == 0 {
		// Пустой файл допустим только при offset=0 и limit=0
		if offset == 0 && limit == 0 {
			return file, nil
		}
		file.Close()
		return nil, ErrUnsupportedFile
	}

	// Размер файла меньше offset - ошибка
	if fileStat.Size() < offset {
		file.Close()
		return nil, ErrOffsetExceedsFileSize
	}

	return file, nil
}
