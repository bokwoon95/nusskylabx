package templateutil

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Txt contains Template Functions that make text more presentable
func Txt(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["Txt_Aan"] = Aan
	funcs["Txt_Itoa"] = Itoa
	funcs["Txt_JoinSlice"] = JoinSlice
	funcs["Txt_JoinSet"] = JoinSet
	funcs["Txt_Title"] = strings.Title
	return funcs
}

// Aan is a Template Function that returns "an" or "a" depending on whether the
// first letter of the string is a vowel. It defaults to "a"
func Aan(input string) string {
	if len(input) > 0 {
		switch unicode.ToLower([]rune(input)[0]) {
		case 'a', 'e', 'i', 'o', 'u':
			return "an"
		}
	}
	return "a"
}

// Itoa is a Template Function that converts a number into a string. If a
// string is provided, the string is simply returned
func Itoa(num interface{}) (string, error) {
	switch v := num.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	default:
		return "", fmt.Errorf("%+v is not a number", num)
	}
}

func JoinSlice(slice []string, sep string) string {
	return strings.Join(slice, sep)
}

func JoinSet(set map[string]bool, sep string) string {
	var builder strings.Builder
	i := 1
	for k := range set {
		switch {
		case i < len(set):
			builder.WriteString(k)
			builder.WriteString(sep)
		default:
			builder.WriteString(k)
		}
		i++
	}
	return builder.String()
}
