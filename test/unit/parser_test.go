package unit

import (
	"os"
	"testing"

	"github.com/budyaya/resty-stress-tester/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVParser(t *testing.T) {
	// 创建临时 CSV 文件
	csvContent := `id,name,email,category
1,John Doe,john@example.com,premium
2,Jane Smith,jane@example.com,standard
3,Bob Wilson,bob@example.com,vip`

	tmpFile, err := os.CreateTemp("", "test*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(csvContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 测试 CSV 解析
	csvParser, err := parser.NewCSVParser(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 3, csvParser.RowCount())

	// 测试获取数据
	data := csvParser.GetData()
	assert.Len(t, data, 3)
	assert.Equal(t, "John Doe", data[0]["name"])
	assert.Equal(t, "standard", data[1]["category"])

	// 测试获取指定行
	row := csvParser.GetRow(0)
	assert.Equal(t, "1", row["id"])
	assert.Equal(t, "premium", row["category"])

	// 测试循环获取
	row = csvParser.GetRow(5) // 应该循环到第 2 行 (5 % 3 = 2)
	assert.Equal(t, "3", row["id"])

	// 测试列头
	headers := csvParser.Headers()
	expectedHeaders := []string{"id", "name", "email", "category"}
	assert.ElementsMatch(t, expectedHeaders, headers)
}

func TestCSVParser_EmptyFile(t *testing.T) {
	// 创建空文件
	tmpFile, err := os.CreateTemp("", "empty*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = parser.NewCSVParser(tmpFile.Name())
	assert.Error(t, err)
}

func TestCSVParser_NonExistentFile(t *testing.T) {
	_, err := parser.NewCSVParser("nonexistent.csv")
	assert.Error(t, err)
}

func TestTemplateParser(t *testing.T) {
	csvParser, err := parser.NewCSVParser("../testdata/sample.csv")
	require.NoError(t, err)

	tmplParser := parser.NewTemplateParser(csvParser)

	// 测试模板处理
	data := map[string]string{
		"id":       "123",
		"name":     "Test User",
		"category": "premium",
	}

	template := "User {{id}}: {{name}} ({{category}})"
	result := tmplParser.Process(template, data)
	expected := "User 123: Test User (premium)"
	assert.Equal(t, expected, result)

	// 测试 JSON 模板处理
	jsonTemplate := `{"id": {{id}}, "name": "{{name}}", "type": "{{category}}"}`
	jsonResult, err := tmplParser.ProcessJSON(jsonTemplate, data)
	require.NoError(t, err)

	// 验证 JSON 结果
	jsonMap, ok := jsonResult.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "123", jsonMap["id"])
	assert.Equal(t, "Test User", jsonMap["name"])

	// 测试 Headers 处理
	headers := map[string]string{
		"Authorization": "Bearer {{id}}",
		"X-User-Name":   "{{name}}",
	}
	processedHeaders := tmplParser.ProcessHeaders(headers, data)
	assert.Equal(t, "Bearer 123", processedHeaders["Authorization"])
	assert.Equal(t, "Test User", processedHeaders["X-User-Name"])

	// 测试 URL 处理
	urlTemplate := "https://api.example.com/users/{{id}}/profile"
	processedURL := tmplParser.ProcessURL(urlTemplate, data)
	assert.Equal(t, "https://api.example.com/users/123/profile", processedURL)
}

func TestTemplateParser_Validation(t *testing.T) {
	tmplParser := parser.NewTemplateParser(nil)

	// 测试模板验证
	validTemplate := "Hello {{name}}"
	err := tmplParser.ValidateTemplate(validTemplate)
	assert.NoError(t, err)

	// 测试不平衡的模板
	invalidTemplate := "Hello {{name"
	err = tmplParser.ValidateTemplate(invalidTemplate)
	assert.Error(t, err)

	// 测试只有闭合标签
	invalidTemplate2 := "Hello name}}"
	err = tmplParser.ValidateTemplate(invalidTemplate2)
	assert.Error(t, err)
}
