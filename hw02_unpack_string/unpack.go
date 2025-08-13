package hw02unpackstring

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	builder := strings.Builder{}

	// Если строка пустая - завершаем ошибки
	if str == "" {
		return builder.String(), nil
	}
	// Если строка не корректная - завершаем с ошибкой
	if !stringIsOk(str) {
		return builder.String(), ErrInvalidString
	}

	runes := []rune(str)
	countRunes := len(runes)
	isStrWithEsc := isStringWithEscaping(str)

	for i := 0; i < countRunes; i++ {
		currRune := runes[i]
		hasNextRune := i+1 <= countRunes-1
		hasAfterNextRune := i+2 <= countRunes-1

		// Если строка с символом '\' и текущий символ '\'
		if isStrWithEsc && currRune == '\\' {
			switch {
			// Следующий и послеследующий символы '\'
			case hasNextRune && hasAfterNextRune && currRune == '\\' && runes[i+1] == '\\' && runes[i+2] == '\\':
				builder.WriteRune(currRune)
				i++
				continue
			// Если следующий символ цифра и послеследующий символ цифра
			case hasNextRune && hasAfterNextRune && runeIsDigit(runes[i+1]) && runeIsDigit(runes[i+2]):
				repeats, err := strconv.Atoi(string(runes[i+2]))
				if err != nil {
					return "", fmt.Errorf("error convert rune '%v' to int", string(runes[i+2]))
				}
				builder.WriteString(strings.Repeat(string(runes[i+1]), repeats))
				i += 2
				continue
			// Если следующий символ '\' и послеследующая руна - цифра
			case hasNextRune && hasAfterNextRune && runes[i+1] == '\\' && runeIsDigit(runes[i+2]):
				repeats, err := strconv.Atoi(string(runes[i+2]))
				if err != nil {
					return "", fmt.Errorf("error convert rune '%v' to int", string(runes[i+2]))
				}
				builder.WriteString(strings.Repeat(string(currRune), repeats))
				i += 2
				continue
			// Если следующий символ цифра
			case hasNextRune && runeIsDigit(runes[i+1]):
				builder.WriteRune(runes[i+1])
				i++
				continue
			}
		}

		if hasNextRune && runeIsDigit(runes[i+1]) {
			repeats, err := strconv.Atoi(string(runes[i+1]))
			if err != nil {
				return "", fmt.Errorf("error convert rune '%v' to int", string(runes[i+2]))
			}
			builder.WriteString(strings.Repeat(string(currRune), repeats))
			i++
			continue
		}

		builder.WriteRune(currRune)
	}
	return builder.String(), nil
}

func stringIsOk(str string) bool {
	runes := []rune(str)
	countRunes := len(runes)
	// Если строка пустая - корректно
	if countRunes == 0 {
		return true
	}
	// Первый символ цифры от 0 до 9
	if runeIsDigit(runes[0]) {
		return false
	}
	// Некорр. экранирование и подряд есть две цифры
	if countRunes > 1 {
		for i := 0; i < countRunes-1; i++ {
			currRune := runes[i]
			nextRune := runes[i+1]
			if currRune == '\\' && !(nextRune == '\\' || runeIsDigit(nextRune)) {
				// Если текущий символ '\' и следующий != '\', и != '0'-'9')
				return false
			}
			if i > 0 && runes[i-1] != '\\' {
				// Если пред. символ - не символ экранирования '\'
				if runeIsDigit(currRune) && runeIsDigit(nextRune) {
					// Если подряд есть две цифры
					return false
				}
			}
		}
	}
	return true
}

func isStringWithEscaping(str string) bool {
	for _, r := range str {
		if r == '\\' {
			return true
		}
	}
	return false
}

func runeIsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
