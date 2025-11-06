package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// CSVParser CSV 解析器
type CSVParser struct {
	data []map[string]string
}

// NewCSVParser 创建 CSV 解析器
func NewCSVParser(filename string) (*CSVParser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
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
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = strings.TrimSpace(records[i][j])
			} else {
				row[header] = ""
			}
		}
		data = append(data, row)
	}

	return &CSVParser{data: data}, nil
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
	return p.data[index%len(p.data)]
}

// RowCount 获取行数
func (p *CSVParser) RowCount() int {
	return len(p.data)
}

// Headers 获取列头
func (p *CSVParser) Headers() []string {
	if len(p.data) == 0 {
		return nil
	}

	headers := make([]string, 0, len(p.data[0]))
	for header := range p.data[0] {
		headers = append(headers, header)
	}
	return headers
}
