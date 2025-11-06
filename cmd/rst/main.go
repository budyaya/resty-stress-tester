package main

import (
	"fmt"
	"os"

	"github.com/budyaya/resty-stress-tester/internal/config"
	"github.com/budyaya/resty-stress-tester/internal/engine"
	"github.com/budyaya/resty-stress-tester/pkg/version"
)

func main() {
	// 加载配置
	cfg, err := config.LoadFromFlags()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("\nUsage:\n")
		printUsage()
		os.Exit(1)
	}

	// 显示测试信息
	fmt.Printf("Resty-Stress-Tester %s\n", version.Version)
	fmt.Printf("Starting stress test...\n")
	fmt.Printf("URL:          %s\n", cfg.URL)
	fmt.Printf("Method:       %s\n", cfg.Method)
	fmt.Printf("Concurrency:  %d\n", cfg.Concurrency)

	if cfg.IsDurationBased() {
		fmt.Printf("Duration:     %v\n", cfg.Duration)
	} else {
		fmt.Printf("Total:        %d\n", cfg.TotalRequests)
	}

	if cfg.CSVFile != "" {
		fmt.Printf("CSV File:     %s\n", cfg.CSVFile)
	}

	if cfg.OutputFile != "" {
		fmt.Printf("Output:       %s\n", cfg.OutputFile)
	}
	fmt.Println()

	// 创建压测引擎
	tester, err := engine.NewStressEngine(cfg)
	if err != nil {
		fmt.Printf("Error creating stress tester: %v\n", err)
		os.Exit(1)
	}

	// 运行压测
	result := tester.Run()

	// 生成报告
	tester.PrintReport()

	// 保存详细报告（如果指定了输出文件）
	if cfg.OutputFile != "" && cfg.ReportFormat != "console" {
		if err := tester.GenerateReport(); err != nil {
			fmt.Printf("Error generating report: %v\n", err)
		} else {
			fmt.Printf("Report saved to: %s\n", cfg.OutputFile)
		}
	}

	// 根据错误率决定退出码
	if result.ShouldFail() {
		fmt.Printf("\n❌ Test failed: High error rate detected (%.1f%%)\n",
			100-result.GetSuccessRate())
		os.Exit(1)
	} else {
		fmt.Printf("\n✅ Test completed successfully\n")
	}
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println(`
Usage:
  rst [flags]

Required Flags:
  -url string        Target URL

Basic Flags:
  -n, -requests int        Total number of requests (default 1000)
  -c, -concurrency int     Number of concurrent workers (default 10)
  -d, -duration duration   Test duration (e.g., 30s, 5m)
  -method string           HTTP method (default "GET")

Request Flags:
  -b, -body string         Request body
  -H, -headers string      Request headers (JSON format)
  -t, -timeout duration    Request timeout (default 30s)
  -keep-alive              Enable keep-alive connections (default true)

Parameterization Flags:
  -csv string              CSV file for parameterization

Output Flags:
  -o, -output string       Output file for detailed logs
  -report string           Report format: console, json, html (default "console")
  -v, -verbose             Enable verbose logging

Other Flags:
  -config string           Config file (JSON or YAML)
  -version, -V             Show version information

Examples:
  # Basic test
  rst -url https://api.example.com/users -n 1000 -c 10

  # Duration-based test
  rst -url https://api.example.com/users -c 50 -d 1m

  # POST request with JSON body
  rst -url https://api.example.com/users -method POST -n 5000 -c 50 \
    -body '{"name":"test"}' -H '{"Content-Type":"application/json"}'

  # CSV parameterization
  rst -url "https://api.example.com/users/{{id}}" -csv users.csv -n 10000 -c 100

  # Save JSON report
  rst -url https://api.example.com/users -n 1000 -c 10 -o results.json -report json
`)
}
