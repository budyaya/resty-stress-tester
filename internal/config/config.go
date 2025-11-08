package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/budyaya/resty-stress-tester/pkg/types"
	"github.com/budyaya/resty-stress-tester/pkg/version"
	"github.com/spf13/viper"
)

// Config 配置管理器
type Config struct {
	*types.StressConfig
	configFile string
}

// LoadFromFlags 从命令行标志加载配置
func LoadFromFlags() (*Config, error) {
	cfg := &Config{
		StressConfig: types.DefaultConfig(),
	}

	// 定义命令行标志
	flag.StringVar(&cfg.URL, "url", "", "Target URL (required)")
	flag.StringVar(&cfg.Method, "method", cfg.Method, "HTTP method (GET, POST, PUT, DELETE, PATCH)")
	flag.IntVar(&cfg.TotalRequests, "n", cfg.TotalRequests, "Total number of requests (shorthand)")
	flag.IntVar(&cfg.TotalRequests, "requests", cfg.TotalRequests, "Total number of requests")
	flag.IntVar(&cfg.Concurrency, "c", cfg.Concurrency, "Number of concurrent workers (shorthand)")
	flag.IntVar(&cfg.Concurrency, "concurrency", cfg.Concurrency, "Number of concurrent workers")
	flag.DurationVar(&cfg.Duration, "d", cfg.Duration, "Test duration (e.g., 30s, 5m) (shorthand)")
	flag.DurationVar(&cfg.Duration, "duration", cfg.Duration, "Test duration (e.g., 30s, 5m)")
	flag.StringVar(&cfg.Body, "b", cfg.Body, "Request body (shorthand)")
	flag.StringVar(&cfg.Body, "body", cfg.Body, "Request body")
	flag.StringVar(&cfg.CSVFile, "csv", cfg.CSVFile, "CSV file for parameterization")
	flag.StringVar(&cfg.OutputFile, "o", cfg.OutputFile, "Output file for detailed logs (shorthand)")
	flag.StringVar(&cfg.OutputFile, "output", cfg.OutputFile, "Output file for detailed logs")
	flag.DurationVar(&cfg.Timeout, "t", cfg.Timeout, "Request timeout (shorthand)")
	flag.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "Request timeout")
	flag.BoolVar(&cfg.KeepAlive, "keep-alive", cfg.KeepAlive, "Enable keep-alive connections")
	flag.BoolVar(&cfg.Verbose, "v", cfg.Verbose, "Enable verbose logging (shorthand)")
	flag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "Enable verbose logging")
	flag.StringVar(&cfg.ReportFormat, "report", cfg.ReportFormat, "Report format (console, json, html)")

	var headers string
	flag.StringVar(&headers, "H", "", "Request headers (JSON format) (shorthand)")
	flag.StringVar(&headers, "headers", "", "Request headers (JSON format)")
	flag.StringVar(&cfg.configFile, "config", "", "Config file (JSON or YAML)")

	// 添加版本标志
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "V", false, "Show version information (shorthand)")

	flag.Parse()

	// 显示版本信息
	if showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	// 如果指定了配置文件，从文件加载
	if cfg.configFile != "" {
		fmt.Printf("Config File:  %s\n", cfg.configFile)
		if err := cfg.loadFromFile(); err != nil {
			return nil, fmt.Errorf("failed to load config file: %v", err)
		}
	}

	// 解析 headers
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &cfg.Headers); err != nil {
			return nil, fmt.Errorf("error parsing headers: %v", err)
		}
	}

	// 验证配置
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadFromFile 从配置文件加载
func (c *Config) loadFromFile() error {
	viper.SetConfigFile(c.configFile)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(c.StressConfig)
}

// validate 验证配置
func (c *Config) validate() error {
	if c.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive")
	}

	if c.Duration == 0 && c.TotalRequests <= 0 {
		return fmt.Errorf("either duration or total requests must be specified")
	}

	if c.Duration > 0 && c.TotalRequests > 0 {
		return fmt.Errorf("cannot specify both duration and total requests")
	}

	// 验证 HTTP 方法
	method := strings.ToUpper(c.Method)
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	if !validMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s", c.Method)
	}

	return nil
}

// IsDurationBased 检查是否基于时长测试
func (c *Config) IsDurationBased() bool {
	return c.Duration > 0
}

// GetTestDescription 获取测试描述
func (c *Config) GetTestDescription() string {
	if c.IsDurationBased() {
		return fmt.Sprintf("%s for %v with %d concurrent workers",
			c.Method, c.Duration, c.Concurrency)
	}
	return fmt.Sprintf("%s %d requests with %d concurrent workers",
		c.Method, c.TotalRequests, c.Concurrency)
}
