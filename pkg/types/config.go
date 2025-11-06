package types

import (
	"time"
)

// StressConfig 压测配置
type StressConfig struct {
	URL           string            `json:"url" yaml:"url"`
	Method        string            `json:"method" yaml:"method"`
	TotalRequests int               `json:"total_requests" yaml:"total_requests"`
	Concurrency   int               `json:"concurrency" yaml:"concurrency"`
	Duration      time.Duration     `json:"duration" yaml:"duration"`
	Headers       map[string]string `json:"headers" yaml:"headers"`
	Body          string            `json:"body" yaml:"body"`
	Timeout       time.Duration     `json:"timeout" yaml:"timeout"`
	KeepAlive     bool              `json:"keep_alive" yaml:"keep_alive"`
	CSVFile       string            `json:"csv_file" yaml:"csv_file"`
	OutputFile    string            `json:"output_file" yaml:"output_file"`
	Verbose       bool              `json:"verbose" yaml:"verbose"`
	ReportFormat  string            `json:"report_format" yaml:"report_format"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *StressConfig {
	return &StressConfig{
		Method:        "GET",
		TotalRequests: 1000,
		Concurrency:   10,
		Timeout:       30 * time.Second,
		KeepAlive:     true,
		ReportFormat:  "console",
	}
}
