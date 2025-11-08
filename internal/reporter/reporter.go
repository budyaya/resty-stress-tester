package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
// ConsoleReport 控制台报告
func (r *StressReporter) ConsoleReport(result *types.StressResult) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("HTTP STRESS TEST REPORT")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Printf("Target URL:          %s\n", r.config.URL)
	fmt.Printf("HTTP Method:         %s\n", r.config.Method)
	fmt.Printf("Concurrency:         %d\n", r.config.Concurrency)

	if r.config.IsDurationBased() {
		fmt.Printf("Test Duration:       %v\n", r.config.Duration)
	} else {
		fmt.Printf("Total Requests:      %d\n", r.config.TotalRequests)
	}

	if r.config.CSVFile != "" {
		fmt.Printf("CSV Data Rows:       %d\n", len(result.DetailedResults))
	}

	fmt.Printf("Actual Duration:     %v\n", result.TotalDuration)
	fmt.Printf("Total Requests:      %d\n", result.TotalRequests)
	fmt.Printf("Successful:          %d\n", result.SuccessfulRequests)
	fmt.Printf("Failed:              %d\n", result.FailedRequests)
	fmt.Printf("Success Rate:        %.2f%%\n", result.GetSuccessRate())

	if result.TotalRequests > 0 {
		fmt.Printf("Requests/sec:        %.2f\n", result.GetRequestsPerSecond())
		fmt.Printf("Avg Response Time:   %v\n", result.GetAverageResponseTime())
		fmt.Printf("Min Response Time:   %v\n", result.MinResponseTime)
		fmt.Printf("Max Response Time:   %v\n", result.MaxResponseTime)
		// 新增分位数统计显示
		fmt.Printf("P50 Response Time:   %v\n", result.P50ResponseTime)
		fmt.Printf("P90 Response Time:   %v\n", result.P90ResponseTime)
		fmt.Printf("P99 Response Time:   %v\n", result.P99ResponseTime)
	}

	// 状态码分布
	r.printStatusCodes(result)

	// 错误分布
	r.printErrorDistribution(result)

	fmt.Println(strings.Repeat("=", 70))

	// 检查是否需要警告
	if result.ShouldFail() {
		fmt.Printf("\n⚠️  Warning: High failure rate (%.1f%%) detected!\n",
			100-result.GetSuccessRate())
	}
}

// printStatusCodes 打印状态码分布
func (r *StressReporter) printStatusCodes(result *types.StressResult) {
	fmt.Println("\nStatus Code Distribution:")

	// 收集状态码并排序
	var codes []int
	result.StatusCodes.Range(func(key, value interface{}) bool {
		codes = append(codes, key.(int))
		return true
	})
	sort.Ints(codes)

	for _, code := range codes {
		if count, ok := result.StatusCodes.Load(code); ok {
			percentage := float64(count.(int64)) / float64(result.TotalRequests) * 100
			fmt.Printf("  %d: %d (%.2f%%)\n", code, count.(int64), percentage)
		}
	}
}

// printErrorDistribution 打印错误分布
func (r *StressReporter) printErrorDistribution(result *types.StressResult) {
	var totalErrors int64
	var errorList []struct {
		error string
		count int64
	}

	result.ErrorCounts.Range(func(key, value interface{}) bool {
		errorList = append(errorList, struct {
			error string
			count int64
		}{key.(string), value.(int64)})
		totalErrors += value.(int64)
		return true
	})

	if totalErrors > 0 {
		fmt.Printf("\nError Distribution (Total: %d):\n", totalErrors)

		// 按错误数量排序
		sort.Slice(errorList, func(i, j int) bool {
			return errorList[i].count > errorList[j].count
		})

		for _, item := range errorList {
			percentage := float64(item.count) / float64(totalErrors) * 100
			// 截断长错误信息
			errorMsg := item.error
			if len(errorMsg) > 80 {
				errorMsg = errorMsg[:77] + "..."
			}
			fmt.Printf("  %s: %d (%.2f%%)\n", errorMsg, item.count, percentage)
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
			"min_response_time":     result.MinResponseTime.String(),
			"max_response_time":     result.MaxResponseTime.String(),
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
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Stress Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .metric { margin: 10px 0; }
        .success { color: green; }
        .error { color: red; }
        .warning { color: orange; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }
    </style>
</head>
<body>
    <div class="header">
        <h1>HTTP Stress Test Report</h1>
        <p>Generated at: {{.Timestamp}}</p>
    </div>
    
    <h2>Test Configuration</h2>
    <table>
        <tr><th>Parameter</th><th>Value</th></tr>
        <tr><td>URL</td><td>{{.Config.URL}}</td></tr>
        <tr><td>Method</td><td>{{.Config.Method}}</td></tr>
        <tr><td>Concurrency</td><td>{{.Config.Concurrency}}</td></tr>
        <tr><td>Total Requests</td><td>{{.Result.TotalRequests}}</td></tr>
    </table>
    
    <h2>Results</h2>
    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Success Rate</td><td class="{{if lt .SuccessRate 90}}error{{else if lt .SuccessRate 95}}warning{{else}}success{{end}}">{{.SuccessRate}}%</td></tr>
        <tr><td>Requests/sec</td><td>{{.RPS}}</td></tr>
        <tr><td>Average Response Time</td><td>{{.AvgResponseTime}}</td></tr>
    </table>
</body>
</html>`

	// 简单的模板替换（实际项目中可以使用 html/template）
	htmlContent := strings.ReplaceAll(htmlTemplate, "{{.Timestamp}}", time.Now().Format(time.RFC3339))
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Config.URL}}", r.config.URL)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Config.Method}}", r.config.Method)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Config.Concurrency}}", fmt.Sprintf("%d", r.config.Concurrency))
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Result.TotalRequests}}", fmt.Sprintf("%d", result.TotalRequests))
	htmlContent = strings.ReplaceAll(htmlContent, "{{.SuccessRate}}", fmt.Sprintf("%.2f", result.GetSuccessRate()))
	htmlContent = strings.ReplaceAll(htmlContent, "{{.RPS}}", fmt.Sprintf("%.2f", result.GetRequestsPerSecond()))
	htmlContent = strings.ReplaceAll(htmlContent, "{{.AvgResponseTime}}", result.GetAverageResponseTime().String())

	if r.config.OutputFile != "" {
		return os.WriteFile(r.config.OutputFile, []byte(htmlContent), 0644)
	}

	fmt.Println(htmlContent)
	return nil
}

// SaveReport 保存报告到文件
func (r *StressReporter) SaveReport(result *types.StressResult, filename string) error {
	return r.generateJSONReport(result)
}
