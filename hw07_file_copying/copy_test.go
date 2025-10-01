package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	sourceFilePath := "testdata/test_data_gol.txt"
	outputFilePath := "testdata/test_data_gol_out.txt"

	t.Run("Positive test. Copy entire file", func(t *testing.T) {
		err := Copy(sourceFilePath, outputFilePath, 0, 0)

		require.NoError(t, err)
		require.FileExists(t, outputFilePath)

		defer os.Remove(outputFilePath)

		sourceFileStat, _ := os.Stat(sourceFilePath)
		sourceFileSize := sourceFileStat.Size()
		outputFileStat, _ := os.Stat(outputFilePath)
		outputFileSize := outputFileStat.Size()
		require.Equal(t, sourceFileSize, outputFileSize, "File sizes should be equal")
	})

	t.Run("Positive test. Copy with limit", func(t *testing.T) {
		var limitBytes int64 = 10

		err := Copy(sourceFilePath, outputFilePath, 0, limitBytes)

		require.NoError(t, err)
		require.FileExists(t, outputFilePath)

		defer os.Remove(outputFilePath)

		outputFileStat, _ := os.Stat(outputFilePath)
		outputFileSize := outputFileStat.Size()
		require.Equal(t, limitBytes, outputFileSize, fmt.Sprintf("Output file size should be %v bytes", limitBytes))
	})

	t.Run("Positive test. Copy with offset and limit", func(t *testing.T) {
		var offsetBytes int64 = 20
		var limitBytes int64 = 10

		err := Copy(sourceFilePath, outputFilePath, offsetBytes, limitBytes)

		require.NoError(t, err)
		require.FileExists(t, outputFilePath)

		defer os.Remove(outputFilePath)

		outputFileStat, _ := os.Stat(outputFilePath)
		outputFileSize := outputFileStat.Size()
		require.Equal(t, limitBytes, outputFileSize, fmt.Sprintf("Output file size should be %v bytes", limitBytes))
	})
	t.Run("Positive test. Copy with large offset and limit", func(t *testing.T) {
		sourceFileStat, _ := os.Stat(sourceFilePath)
		sourceFileSize := sourceFileStat.Size()
		offsetBytes := sourceFileSize - 10
		var limitBytes int64 = 100

		err := Copy(sourceFilePath, outputFilePath, offsetBytes, limitBytes)

		require.NoError(t, err)
		require.FileExists(t, outputFilePath)

		defer os.Remove(outputFilePath)

		outputFileStat, _ := os.Stat(outputFilePath)
		outputFileSize := outputFileStat.Size()
		expectedFileSize := min(limitBytes, sourceFileSize-offsetBytes)
		require.Equal(t, expectedFileSize, outputFileSize,
			fmt.Sprintf("Output file size should be %v bytes", expectedFileSize))
	})

	t.Run("Negative test. Empty source or output file path", func(t *testing.T) {
		err := Copy("", outputFilePath, 0, 0)
		require.ErrorIs(t, err, ErrOffPaths)
		err = Copy(sourceFilePath, "", 0, 0)
		require.ErrorIs(t, err, ErrOffPaths)
	})

	t.Run("Negative test. Negative offset or limit", func(t *testing.T) {
		err := Copy(sourceFilePath, outputFilePath, -1, 0)
		require.ErrorIs(t, err, ErrNegativeLimitOrOffset)
		err = Copy(sourceFilePath, outputFilePath, 0, -1)
		require.ErrorIs(t, err, ErrNegativeLimitOrOffset)
	})

	t.Run("Negative test. Оffset exceeds file size", func(t *testing.T) {
		sourceFileStat, _ := os.Stat(sourceFilePath)
		sourceFileSize := sourceFileStat.Size()
		offset := sourceFileSize + 1
		err := Copy(sourceFilePath, outputFilePath, offset, 0)
		require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
	})

	t.Run("Negative test. Unsupported file", func(t *testing.T) {
		// Создаем временный каталог (вместо source файла)
		tmpDir := t.TempDir()
		err := Copy(tmpDir, outputFilePath, 0, 0)
		require.ErrorIs(t, err, ErrUnsupportedFile)
		require.Equal(t, ErrUnsupportedFile, err)
	})

	t.Run("Negative test. Copy from non-existent file", func(t *testing.T) {
		err := Copy("nonexistent.txt", outputFilePath, 0, 0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error opening file")
	})
}
