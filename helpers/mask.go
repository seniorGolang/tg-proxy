package helpers

import (
	"strings"
)

const (
	maskPrefixLength = 4
	maskSuffixLength = 4
	maskChar         = "*"
	minLengthToMask  = 8
)

// MaskSensitive: первые и последние 4 символа видимы, остальное — звёздочки; при длине < 8 маскирует полностью.
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

func MaskToken(token string) (masked string) {

	return MaskSensitive(token)
}

func MaskKey(key string) (masked string) {

	return MaskSensitive(key)
}

func MaskPassword(password string) (masked string) {

	if password == "" {
		return ""
	}

	return strings.Repeat(maskChar, len(password))
}
