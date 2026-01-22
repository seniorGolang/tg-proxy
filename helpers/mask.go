package helpers

import (
	"strings"
)

const (
	// maskPrefixLength количество символов в начале строки, которые не маскируются
	maskPrefixLength = 4
	// maskSuffixLength количество символов в конце строки, которые не маскируются
	maskSuffixLength = 4
	// maskChar символ для маскирования
	maskChar = "*"
	// minLengthToMask минимальная длина строки для маскирования
	minLengthToMask = 8
)

// MaskSensitive маскирует чувствительную информацию
// Показывает первые 4 и последние 4 символа, остальное заменяет на *
// Если строка короче 8 символов, маскирует полностью
func MaskSensitive(value string) (masked string) {

	if value == "" {
		return ""
	}

	length := len(value)
	if length < minLengthToMask {
		return strings.Repeat(maskChar, length)
	}

	if length <= maskPrefixLength+maskSuffixLength {
		return strings.Repeat(maskChar, length)
	}

	prefix := value[:maskPrefixLength]
	suffix := value[length-maskSuffixLength:]
	maskedLength := length - maskPrefixLength - maskSuffixLength

	return prefix + strings.Repeat(maskChar, maskedLength) + suffix
}

// MaskToken маскирует токен (специальная обработка для токенов)
func MaskToken(token string) (masked string) {

	return MaskSensitive(token)
}

// MaskKey маскирует ключ (специальная обработка для ключей)
func MaskKey(key string) (masked string) {

	return MaskSensitive(key)
}

// MaskPassword маскирует пароль (полностью маскируется)
func MaskPassword(password string) (masked string) {

	if password == "" {
		return ""
	}

	return strings.Repeat(maskChar, len(password))
}
