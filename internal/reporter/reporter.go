package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/budyaya/resty-stress-tester/internal/config"
	"github.com/budyaya/resty-stress-tester/pkg/types"
)

// Reporter 报告生成器接口
type Reporter interface {
	GenerateReport(result *types.StressResult) error
	ConsoleReport(result *types.StressResult)
	SaveReport(result *types.StressResult, filename string) error
}

// StressReporter 压测报告生成器
type StressReporter struct {
	config *config.Config
}

// NewReporter 创建报告生成器
func NewReporter(cfg *config.Config) *StressReporter {
	return &StressReporter{
		config: cfg,
	}
}

// GenerateReport 生成报告
func (r *StressReporter) GenerateReport(result *types.StressResult) error {
	switch r.config.ReportFormat {
	case "json":
		return r.generateJSONReport(result)
	case "html":
		return r.generateHTMLReport(result)
	default:
		r.ConsoleReport(result)
		return nil
	}
}

// ConsoleReport 控制台报告
func (r *StressReporter) ConsoleReport(result *types.StressResult) {
	var buf strings.Builder
	buf.WriteString("\n" + strings.Repeat("=", 70) + "\n")
	buf.WriteString("HTTP STRESS TEST REPORT\n")
	buf.WriteString(strings.Repeat("=", 70) + "\n")

	buf.WriteString(fmt.Sprintf("Target URL:          %s\n", r.config.URL))
	buf.WriteString(fmt.Sprintf("HTTP Method:         %s\n", r.config.Method))
	buf.WriteString(fmt.Sprintf("Concurrency:         %d\n", r.config.Concurrency))

	if r.config.IsDurationBased() {
		buf.WriteString(fmt.Sprintf("Test Duration:       %v\n", r.config.Duration))
	} else {
		buf.WriteString(fmt.Sprintf("Total Requests:      %d\n", r.config.TotalRequests))
	}

	if r.config.CSVFile != "" {
		buf.WriteString(fmt.Sprintf("CSV Data Rows:       %d\n", len(result.DetailedResults)))
	}

	buf.WriteString(fmt.Sprintf("Actual Duration:     %v\n", result.TotalDuration))
	buf.WriteString(fmt.Sprintf("Total Requests:      %d\n", result.TotalRequests))
	buf.WriteString(fmt.Sprintf("Successful:          %d\n", result.SuccessfulRequests))
	buf.WriteString(fmt.Sprintf("Failed:              %d\n", result.FailedRequests))
	buf.WriteString(fmt.Sprintf("Success Rate:        %.2f%%\n", result.GetSuccessRate()))

	if result.TotalRequests > 0 {
		buf.WriteString(fmt.Sprintf("Requests/sec:        %.2f\n", result.GetRequestsPerSecond()))
		buf.WriteString(fmt.Sprintf("Avg Response Time:   %v\n", result.GetAverageResponseTime()))
		buf.WriteString(fmt.Sprintf("Min Response Time:   %v\n", result.GetMinResponseTime()))
		buf.WriteString(fmt.Sprintf("Max Response Time:   %v\n", result.GetMaxResponseTime()))
		// 新增分位数统计显示
		buf.WriteString(fmt.Sprintf("P50 Response Time:   %v\n", result.P50ResponseTime))
		buf.WriteString(fmt.Sprintf("P90 Response Time:   %v\n", result.P90ResponseTime))
		buf.WriteString(fmt.Sprintf("P99 Response Time:   %v\n", result.P99ResponseTime))
	}

	// 状态码分布
	r.writeStatusCodes(&buf, result)

	// 错误分布
	r.writeErrorDistribution(&buf, result)

	buf.WriteString(strings.Repeat("=", 70) + "\n")

	// 检查是否需要警告
	if result.ShouldFail() {
		buf.WriteString(fmt.Sprintf("\n⚠️  Warning: High failure rate (%.1f%%) detected!\n",
			100-result.GetSuccessRate()))
	}

	fmt.Print(buf.String())
}

// writeStatusCodes 写入状态码分布
func (r *StressReporter) writeStatusCodes(buf *strings.Builder, result *types.StressResult) {
	buf.WriteString("\nStatus Code Distribution:\n")

	// 使用更高效的方式收集状态码
	codes := result.GetSortedStatusCodes()

	for _, code := range codes {
		count := result.GetStatusCodeCount(code)
		percentage := float64(count) / float64(result.TotalRequests) * 100
		buf.WriteString(fmt.Sprintf("  %d: %d (%.2f%%)\n", code, count, percentage))
	}
}

// writeErrorDistribution 写入错误分布
func (r *StressReporter) writeErrorDistribution(buf *strings.Builder, result *types.StressResult) {
	errorList, totalErrors := result.GetSortedErrors()

	if totalErrors > 0 {
		buf.WriteString(fmt.Sprintf("\nError Distribution (Total: %d):\n", totalErrors))

		for _, item := range errorList {
			percentage := float64(item.Count) / float64(totalErrors) * 100
			// 截断长错误信息
			errorMsg := item.Error
			if len(errorMsg) > 80 {
				errorMsg = errorMsg[:77] + "..."
			}
			buf.WriteString(fmt.Sprintf("  %s: %d (%.2f%%)\n", errorMsg, item.Count, percentage))
		}
	}
}

// generateJSONReport 生成 JSON 报告
func (r *StressReporter) generateJSONReport(result *types.StressResult) error {
	report := struct {
		Config  *config.Config         `json:"config"`
		Result  *types.StressResult    `json:"result"`
		Summary map[string]interface{} `json:"summary"`
	}{
		Config: r.config,
		Result: result,
		Summary: map[string]interface{}{
			"requests_per_second":   result.GetRequestsPerSecond(),
			"success_rate":          result.GetSuccessRate(),
			"average_response_time": result.GetAverageResponseTime().String(),
			"min_response_time":     result.GetMinResponseTime().String(),
			"max_response_time":     result.GetMaxResponseTime().String(),
			"p50_response_time":     result.P50ResponseTime.String(),
			"p90_response_time":     result.P90ResponseTime.String(),
			"p99_response_time":     result.P99ResponseTime.String(),
		},
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if r.config.OutputFile != "" {
		return os.WriteFile(r.config.OutputFile, jsonData, 0644)
	}

	fmt.Println(string(jsonData))
	return nil
}

// generateHTMLReport 生成 HTML 报告
func (r *StressReporter) generateHTMLReport(result *types.StressResult) error {

	// 使用 strings.Builder 进行模板替换
	var htmlContent strings.Builder
	htmlContent.WriteString("<!DOCTYPE html>\n<html>\n<head>\n    <title>Stress Test Report</title>\n    <style>\n        body { font-family: Arial, sans-serif; margin: 20px; }\n        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }\n        .metric { margin: 10px 0; }\n        .success { color: green; }\n        .error { color: red; }\n        .warning { color: orange; }\n        table { width: 100%; border-collapse: collapse; }\n        th, td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }\n    </style>\n</head>\n<body>\n    <div class=\"header\">\n        <h1>HTTP Stress Test Report</h1>\n        <p>Generated at: ")
	htmlContent.WriteString(time.Now().Format(time.RFC3339))
	htmlContent.WriteString("</p>\n    </div>\n    \n    <h2>Test Configuration</h2>\n    <table>\n        <tr><th>Parameter</th><th>Value</th></tr>\n        <tr><td>URL</td><td>")
	htmlContent.WriteString(r.config.URL)
	htmlContent.WriteString("</td></tr>\n        <tr><td>Method</td><td>")
	htmlContent.WriteString(r.config.Method)
	htmlContent.WriteString("</td></tr>\n        <tr><td>Concurrency</td><td>")
	htmlContent.WriteString(fmt.Sprintf("%d", r.config.Concurrency))
	htmlContent.WriteString("</td></tr>\n        <tr><td>Total Requests</td><td>")
	htmlContent.WriteString(fmt.Sprintf("%d", result.TotalRequests))
	htmlContent.WriteString("</td></tr>\n    </table>\n    \n    <h2>Results</h2>\n    <table>\n        <tr><th>Metric</th><th>Value</th></tr>\n        <tr><td>Success Rate</td><td class=\"")

	successRate := result.GetSuccessRate()
	if successRate < 90 {
		htmlContent.WriteString("error")
	} else if successRate < 95 {
		htmlContent.WriteString("warning")
	} else {
		htmlContent.WriteString("success")
	}

	htmlContent.WriteString("\">")
	htmlContent.WriteString(fmt.Sprintf("%.2f%%", successRate))
	htmlContent.WriteString("</td></tr>\n        <tr><td>Requests/sec</td><td>")
	htmlContent.WriteString(fmt.Sprintf("%.2f", result.GetRequestsPerSecond()))
	htmlContent.WriteString("</td></tr>\n        <tr><td>Average Response Time</td><td>")
	htmlContent.WriteString(result.GetAverageResponseTime().String())
	htmlContent.WriteString("</td></tr>\n    </table>\n</body>\n</html>")

	if r.config.OutputFile != "" {
		return os.WriteFile(r.config.OutputFile, []byte(htmlContent.String()), 0644)
	}

	fmt.Println(htmlContent.String())
	return nil
}

// SaveReport 保存报告到文件
func (r *StressReporter) SaveReport(result *types.StressResult, filename string) error {
	return r.generateJSONReport(result)
}
