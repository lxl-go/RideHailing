package validator

import (
	"regexp"
)

// IsValidPhone 校验手机号（11 位，1 开头）
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// IsValidIDCard 简单身份证号格式校验（18 位）
var idCardRegex = regexp.MustCompile(`^\d{17}[\dXx]$`)

func IsValidIDCard(card string) bool {
	return idCardRegex.MatchString(card)
}

// NormalizePage 分页参数规范化（对标文档 4.5 节分页规范）
func NormalizePage(page, pageSize int32) (p, size int32) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
