package locale

import "strings"

const (
	// TagZH 简体中文标签，与常见前端 i18n 资源对齐。
	TagZH = "zh-CN"
	// TagEN 英语标签（缺省或非中文首选时）。
	TagEN = "en"
)

// NormalizeFromAcceptLanguage 根据 HTTP Accept-Language 头值归约为 TagZH 或 TagEN：
// 忽略大小写，若头字符串以 zh 为前缀则视为中文，否则为英文；空串视为英文。
func NormalizeFromAcceptLanguage(acceptLanguage string) string {
	lang := acceptLanguage
	if lang == "" {
		return TagEN
	}
	if strings.HasPrefix(strings.ToLower(lang), "zh") {
		return TagZH
	}
	return TagEN
}
