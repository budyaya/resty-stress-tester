package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger 日志记录器
type Logger struct {
	verbose bool
	logger  *log.Logger
	file    *os.File
}

// NewLogger 创建日志记录器
func NewLogger(verbose bool, logFile string) (*Logger, error) {
	logger := &Logger{
		verbose: verbose,
		logger:  log.New(os.Stdout, "", log.LstdFlags),
	}

	if logFile != "" {
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %v", err)
		}

		file, err := os.Create(logFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create log file: %v", err)
		}

		logger.file = file
		logger.logger = log.New(file, "", log.LstdFlags)
	}

	return logger, nil
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

// Progress 显示进度
func (l *Logger) Progress(current, total int64, startTime time.Time) {
	if !l.verbose {
		return
	}

	elapsed := time.Since(startTime)
	percent := float64(current) / float64(total) * 100
	rps := float64(current) / elapsed.Seconds()

	fmt.Printf("\rProgress: %d/%d (%.1f%%) - %.1f req/sec",
		current, total, percent, rps)
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
