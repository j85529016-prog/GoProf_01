package hw10programoptimization

import (
	"bufio"
	"io"
	"strings"
)

type DomainStat map[string]int

// extractEmail быстро извлекает email из строки вида {..."Email":"user@domain.ru",...}.
func extractEmail(s string) string {
	// Быстро находим `"Email":"`
	const prefix = `"Email":"`
	i := strings.Index(s, prefix)
	if i == -1 {
		return ""
	}
	i += len(prefix)

	// Находим конец значения (закрывающая кавычка, не экранированная)
	for j := i; j < len(s); j++ {
		if s[j] == '"' {
			if j == 0 || s[j-1] != '\\' {
				return s[i:j]
			}
		}
	}
	return ""
}

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	scanner := bufio.NewScanner(r)
	domainStatResult := make(DomainStat)
	domainSuffix := "." + domain
	suffixLen := len(domainSuffix)

	for scanner.Scan() {
		// Cразу преобразуем строку в string
		line := scanner.Text()
		if line == "" {
			continue
		}

		email := extractEmail(line)
		if email == "" {
			continue
		}

		if len(email) < suffixLen {
			continue
		}

		// Быстрое сравнение без аллокаций
		if !strings.EqualFold(email[len(email)-suffixLen:], domainSuffix) {
			continue
		}

		// Быстрый split по @
		if i := strings.IndexByte(email, '@'); i != -1 {
			domainPart := strings.ToLower(email[i+1:])
			domainStatResult[domainPart]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return domainStatResult, nil
}
