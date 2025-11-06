package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TemplateParser 模板解析器
type TemplateParser struct {
	csvParser *CSVParser
}

// NewTemplateParser 创建模板解析器
func NewTemplateParser(csvParser *CSVParser) *TemplateParser {
	return &TemplateParser{
		csvParser: csvParser,
	}
}

// Process 处理模板字符串
func (p *TemplateParser) Process(template string, data map[string]string) string {
	if data == nil {
		return template
	}

	result := template
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// ProcessJSON 处理 JSON 模板
func (p *TemplateParser) ProcessJSON(template string, data map[string]string) (interface{}, error) {
	processed := p.Process(template, data)

	// 尝试解析为 JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(processed), &jsonData); err != nil {
		// 如果不是有效的 JSON，返回原始字符串
		return processed, nil
	}

	return jsonData, nil
}

// ProcessURL 处理 URL 模板
func (p *TemplateParser) ProcessURL(urlTemplate string, data map[string]string) string {
	return p.Process(urlTemplate, data)
}

// ProcessHeaders 处理 Headers 模板
func (p *TemplateParser) ProcessHeaders(headers map[string]string, data map[string]string) map[string]string {
	if data == nil {
		return headers
	}

	processed := make(map[string]string)
	for key, value := range headers {
		processed[key] = p.Process(value, data)
	}
	return processed
}

// ValidateTemplate 验证模板
func (p *TemplateParser) ValidateTemplate(template string) error {
	// 检查未闭合的模板标签
	openCount := strings.Count(template, "{{")
	closeCount := strings.Count(template, "}}")

	if openCount != closeCount {
		return fmt.Errorf("unbalanced template tags: %d opening vs %d closing", openCount, closeCount)
	}

	return nil
}

// GetAvailableVariables 获取可用变量
func (p *TemplateParser) GetAvailableVariables() []string {
	if p.csvParser == nil {
		return nil
	}
	return p.csvParser.Headers()
}
