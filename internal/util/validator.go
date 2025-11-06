package util

import (
	"fmt"
	"net/url"
	"strings"
)

// Validator 验证器
type Validator struct{}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateURL 验证 URL
func (v *Validator) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must contain a host")
	}

	return nil
}

// ValidateMethod 验证 HTTP 方法
func (v *Validator) ValidateMethod(method string) error {
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	method = strings.ToUpper(method)
	if !validMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}

	return nil
}

// ValidateConcurrency 验证并发数
func (v *Validator) ValidateConcurrency(concurrency int) error {
	if concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive")
	}
	if concurrency > 10000 {
		return fmt.Errorf("concurrency too high: %d (max: 10000)", concurrency)
	}
	return nil
}

// ValidateRequests 验证请求数
func (v *Validator) ValidateRequests(requests int) error {
	if requests < 0 {
		return fmt.Errorf("requests cannot be negative")
	}
	if requests > 10000000 {
		return fmt.Errorf("requests too high: %d (max: 10,000,000)", requests)
	}
	return nil
}
