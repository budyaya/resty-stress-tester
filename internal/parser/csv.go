package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// CSVParser CSV 解析器
type CSVParser struct {
	data     []map[string]string
	headers  []string
	rowCount int
}

// NewCSVParser 创建 CSV 解析器
func NewCSVParser(filename string) (*CSVParser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	// 使用带缓存的reader提高大文件读取性能
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允许可变字段数
	reader.LazyQuotes = true    // 允许宽松的引号处理

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	headers := make([]string, len(records[0]))
	for i, header := range records[0] {
		headers[i] = strings.TrimSpace(header)
	}

	data := make([]map[string]string, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		row := make(map[string]string, len(headers))
		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = strings.TrimSpace(records[i][j])
			} else {
				row[header] = ""
			}
		}
		data = append(data, row)
	}

	return &CSVParser{
		data:     data,
		headers:  headers,
		rowCount: len(data),
	}, nil
}

// GetData 获取所有数据
func (p *CSVParser) GetData() []map[string]string {
	return p.data
}

// GetRow 获取指定行数据
func (p *CSVParser) GetRow(index int) map[string]string {
	if len(p.data) == 0 {
		return nil
	}

	// 使用取模运算实现循环读取
	if p.rowCount > 0 {
		return p.data[index%p.rowCount]
	}
	return p.data[index%len(p.data)]
}

// RowCount 获取行数
func (p *CSVParser) RowCount() int {
	if p.rowCount > 0 {
		return p.rowCount
	}
	return len(p.data)
}

// Headers 获取列头
func (p *CSVParser) Headers() []string {
	return p.headers
}

// Close 关闭解析器（清理资源）
func (p *CSVParser) Close() error {
	// 目前没有需要特别清理的资源
	// 但如果使用了内存映射等，需要在这里清理
	p.data = nil
	p.headers = nil
	return nil
}
