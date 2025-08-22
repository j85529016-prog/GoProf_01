package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

func Top10(str string) []string {
	if str == "" {
		return []string{}
	}

	// Сплитуем строку в слайс слов
	words := strings.Fields(str)

	// Формируем шаблон для очистки слов от знаков препинания и пробелов по краям слова
	cleanPattern := regexp.MustCompile(`^[\p{P}\s]+|[\p{P}\s]+$`)

	// Заполняем мапу частоты слов, где ключ - слово, значение - частота
	mapWordFreq := make(map[string]int, len(words))
	for _, word := range words {
		// Очищаем слово и приводим к нижнему регистру
		cleanWord := cleanPattern.ReplaceAllString(word, "")
		cleanWord = strings.ToLower(cleanWord)
		// Если очищенное слово получилось "" или "-", то пропускаем его
		if cleanWord == "" || cleanWord == "-" {
			continue
		}
		mapWordFreq[cleanWord]++
	}

	// Создаем структуру слово-частота, и заполняем слайс таких структур
	type wordFreq struct {
		word      string
		frequency int
	}
	slWordFreq := make([]wordFreq, 0, len(mapWordFreq))
	for k, v := range mapWordFreq {
		slWordFreq = append(slWordFreq, wordFreq{k, v})
	}

	// Сортируем слайс структур
	sort.Slice(slWordFreq, func(i, j int) bool {
		// Если есть слова с одинаковой частотой - сортируем по словам ASC
		if slWordFreq[i].frequency == slWordFreq[j].frequency {
			return slWordFreq[i].word < slWordFreq[j].word
		}
		// Сортируем по частоте DESC
		return slWordFreq[i].frequency > slWordFreq[j].frequency
	})

	// Заполняем итоговый результат
	result := make([]string, 0, len(slWordFreq))
	for i, v := range slWordFreq {
		// Прерываем если top10 уже набран
		if i >= 10 {
			return result
		}
		result = append(result, v.word)
	}

	return result
}
