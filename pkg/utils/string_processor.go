package utils

import (
	"strings"
	"unicode"
)

type StringType struct {
	TitleCase      string
	CamelCase      string
	SnakeCaseLower string
	SnakeCaseUpper string
	kebabCase      string
}

func toCamelCase(str string) string {
	runes := []rune(str)
	var result []rune
	isFirst := true
	nextUpper := false

	for _, r := range runes {
		if r == '_' || unicode.IsSpace(r) {
			nextUpper = true
			continue
		}

		if isFirst {
			result = append(result, unicode.ToLower(r))
			isFirst = false
		} else if nextUpper {
			result = append(result, unicode.ToUpper(r))
			nextUpper = false
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func toTitleCase(str string) string {
	runes := []rune(str)
	var result []rune
	nextUpper := true

	for i, r := range runes {
		if r == '_' || unicode.IsSpace(r) {
			nextUpper = true
			continue
		}

		if nextUpper {
			result = append(result, unicode.ToUpper(r))
			nextUpper = false
		} else {
			result = append(result, runes[i])
		}
	}
	return string(result)
}

func toSnakeCaseLower(str string) string {
	runes := []rune(str)
	var result []rune

	for i, r := range runes {
		if i > 0 && (unicode.IsUpper(r) || unicode.IsSpace(runes[i-1])) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return strings.Trim(string(result), "_")
}

func ProcessString(input string) StringType {
	processedString := StringType{
		TitleCase:      toTitleCase(input),
		CamelCase:      toCamelCase(input),
		SnakeCaseLower: toSnakeCaseLower(input),
		SnakeCaseUpper: strings.ToUpper(toSnakeCaseLower(input)),
		kebabCase:      toKebabCase(input),
	}

	return processedString
}

func toKebabCase(str string) string {
	runes := []rune(str)
	var result []rune

	for i, r := range runes {
		if i > 0 && (unicode.IsUpper(r) || unicode.IsSpace(runes[i-1])) {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return strings.Trim(string(result), "-")
}
