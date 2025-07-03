//go:generate go run github.com/abice/go-enum --names --values --marshal --sql --flag --nocase

package types

import (
	"fmt"
	"strings"
)

// Language represents supported languages for documentation generation
// ENUM(English, Russian)
type Language string

// TODO(dm.a.kudryavtsev): кажется можно коды объединить в один enum с Language

// Code returns the language code (for CLI flags and config)
func (l Language) Code() string {
	switch l {
	case LanguageEnglish:
		return "en"
	case LanguageRussian:
		return "ru"
	default:
		return "en"
	}
}

// ParseLanguageWithCode parses a language from string (accepts both full name and code)
// This extends the generated ParseLanguage function to also handle language codes
func ParseLanguageWithCode(s string) (Language, error) {
	s = strings.TrimSpace(s)

	// First try the generated ParseLanguage (handles full names)
	if lang, err := ParseLanguage(s); err == nil {
		return lang, nil
	}

	// Try by language code
	switch strings.ToLower(s) {
	case "en":
		return LanguageEnglish, nil
	case "ru":
		return LanguageRussian, nil
	}

	return "", fmt.Errorf("unsupported language: %s (supported: English/en, Russian/ru)", s)
}

// AllLanguageCodes returns all supported language codes
func AllLanguageCodes() []string {
	langs := LanguageValues()
	codes := make([]string, len(langs))
	for i, lang := range langs {
		codes[i] = lang.Code()
	}
	return codes
}
