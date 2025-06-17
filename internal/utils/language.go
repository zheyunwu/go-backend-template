package utils

// Language is a type for language codes.
type Language string

const (
	LanguageZh Language = "zh" // Chinese
	LanguageEn Language = "en" // English
	LanguageDe Language = "de" // German
)

// IsValid checks if the language code is allowed.
func (l Language) IsValid() bool {
	switch l {
	case LanguageZh, LanguageEn, LanguageDe:
		return true
	}
	return false
}

var LanguageMap = map[string]string{
	"zh": "中文",
	"en": "English",
	"de": "Deutsch",
}
