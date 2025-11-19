package utils

import (
	"regexp"
	"unicode"
)

// ValidateUsername 验证用户名（中文、英文、数字）
func ValidateUsername(username string) bool {
	// 长度检查
	if len(username) < 1 || len(username) > 50 {
		return false
	}

	// 检查是否只包含中文、英文、数字
	for _, r := range username {
		if !(unicode.Is(unicode.Han, r) || unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

// ValidatePassword 验证密码格式
func ValidatePassword(password string) bool {
	// 长度检查
	if len(password) < 3 || len(password) > 50 {
		return false
	}

	// 允许英文、数字、特殊字符
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]*$`, password)
	return matched
}
