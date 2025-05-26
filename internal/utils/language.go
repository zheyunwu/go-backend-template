package utils

// Language 是一个语言代码类型
type Language string

const (
	LanguageZh Language = "zh"
	LanguageEn Language = "en"
	LanguageDe Language = "de"
)

// IsValid 用来校验是否是允许的语言
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
