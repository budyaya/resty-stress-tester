package types

import (
	"time"
)

// StressConfig 压测配置
type StressConfig struct {
	URL           string            `mapstructure:"url" json:"url" yaml:"url"`
	Method        string            `mapstructure:"method" json:"method" yaml:"method"`
	TotalRequests int               `mapstructure:"total_requests" json:"total_requests" yaml:"total_requests"`
	Concurrency   int               `mapstructure:"concurrency" json:"concurrency" yaml:"concurrency"`
	Duration      time.Duration     `mapstructure:"duration" json:"duration" yaml:"duration"`
	Headers       map[string]string `mapstructure:"headers" json:"headers" yaml:"headers"`
	Body          string            `mapstructure:"body" json:"body" yaml:"body"`
	Timeout       time.Duration     `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	KeepAlive     bool              `mapstructure:"keep_alive" json:"keep_alive" yaml:"keep_alive"`
	CSVFile       string            `mapstructure:"csv_file" json:"csv_file" yaml:"csv_file"`
	OutputFile    string            `mapstructure:"output_file" json:"output_file" yaml:"output_file"`
	Verbose       bool              `mapstructure:"verbose" json:"verbose" yaml:"verbose"`
	LogFile       string            `mapstructure:"log_file" json:"log_file" yaml:"log_file"`
	ReportFormat  string            `mapstructure:"report_format" json:"report_format" yaml:"report_format"`
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
