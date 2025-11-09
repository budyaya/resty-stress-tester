package util

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// 默认缓冲区大小
	defaultBufferSize = 8192
	// 默认刷新间隔
	defaultFlushInterval = 5 * time.Second
	// 默认日志文件大小限制 (100MB)
	defaultMaxFileSize = 100 * 1024 * 1024
)

// Logger 日志记录器
type Logger struct {
	verbose        bool
	logger         *log.Logger
	file           *os.File
	writer         *bufio.Writer
	mu             sync.RWMutex
	lastLineLength int

	// 文件轮转相关
	logFilePath string
	maxFileSize int64
	currentSize int64

	// 异步日志支持
	asyncQueue chan string
	asyncWg    sync.WaitGroup
	asyncStop  chan struct{}

	// 定期刷新
	flushTicker *time.Ticker
	flushStop   chan struct{}
	flushWg     sync.WaitGroup
}

// LoggerOptions 日志配置选项
type LoggerOptions struct {
	Verbose       bool
	LogFile       string
	BufferSize    int
	FlushInterval time.Duration
	MaxFileSize   int64
}

// NewLogger 创建日志记录器
func NewLogger(verbose bool, logFile string) (*Logger, error) {
	opts := LoggerOptions{
		Verbose:       verbose,
		LogFile:       logFile,
		BufferSize:    defaultBufferSize,
		FlushInterval: defaultFlushInterval,
		MaxFileSize:   defaultMaxFileSize,
	}
	return NewLoggerWithOptions(opts)
}

// NewLoggerWithOptions 使用配置选项创建日志记录器
func NewLoggerWithOptions(opts LoggerOptions) (*Logger, error) {
	logger := &Logger{
		verbose:     opts.Verbose,
		asyncQueue:  make(chan string, 1000),
		asyncStop:   make(chan struct{}),
		flushStop:   make(chan struct{}),
		logFilePath: opts.LogFile,
		maxFileSize: opts.MaxFileSize,
	}

	// 初始化日志输出
	if opts.LogFile != "" {
		if err := logger.initFileLogging(opts.LogFile, opts.BufferSize); err != nil {
			return nil, err
		}
	} else {
		logger.logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// 启动异步日志处理
	logger.asyncWg.Add(1)
	go logger.processAsyncLogs()

	// 启动定期刷新
	if logger.writer != nil {
		logger.flushWg.Add(1)
		logger.flushTicker = time.NewTicker(opts.FlushInterval)
		go logger.periodicFlush()
	}

	return logger, nil
}

// initFileLogging 初始化文件日志
func (l *Logger) initFileLogging(logFile string, bufferSize int) error {
	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// 获取当前文件大小
	if info, err := file.Stat(); err == nil {
		atomic.StoreInt64(&l.currentSize, info.Size())
	}

	l.file = file
	l.writer = bufio.NewWriterSize(file, bufferSize)
	l.logger = log.New(l.writer, "", log.LstdFlags)

	return nil
}

// processAsyncLogs 处理异步日志
func (l *Logger) processAsyncLogs() {
	defer l.asyncWg.Done()

	for {
		select {
		case msg := <-l.asyncQueue:
			l.writeLog(msg)
		case <-l.asyncStop:
			// 处理剩余日志
			for {
				select {
				case msg := <-l.asyncQueue:
					l.writeLog(msg)
				default:
					return
				}
			}
		}
	}
}

// writeLog 写入日志（线程安全）
func (l *Logger) writeLog(msg string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.logger != nil {
		l.logger.Print(msg)

		// 更新文件大小
		if l.writer != nil {
			atomic.AddInt64(&l.currentSize, int64(len(msg)+1)) // +1 换行符

			// 检查是否需要轮转
			if atomic.LoadInt64(&l.currentSize) > l.maxFileSize {
				go l.rotateLog()
			}
		}
	}
}

// periodicFlush 定期刷新缓冲区
func (l *Logger) periodicFlush() {
	defer l.flushWg.Done()

	for {
		select {
		case <-l.flushTicker.C:
			l.flush()
		case <-l.flushStop:
			return
		}
	}
}

// flush 刷新缓冲区（线程安全）
func (l *Logger) flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer != nil {
		if err := l.writer.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush log buffer: %v\n", err)
		}
	}
}

// rotateLog 日志轮转
func (l *Logger) rotateLog() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 关闭当前文件
	if l.file != nil {
		if l.writer != nil {
			l.writer.Flush()
		}
		l.file.Close()
	}

	// 重命名当前日志文件
	if l.logFilePath != "" {
		timestamp := time.Now().Format("20060102-150405")
		backupPath := l.logFilePath + "." + timestamp
		os.Rename(l.logFilePath, backupPath)
	}

	// 创建新文件
	if err := l.initFileLogging(l.logFilePath, l.writer.Size()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
		// 降级到标准输出
		l.logger = log.New(os.Stdout, "", log.LstdFlags)
		l.writer = nil
		l.file = nil
	}
}

// logAsync 异步记录日志
func (l *Logger) logAsync(level, format string, args ...interface{}) {
	msg := fmt.Sprintf("[%s] "+format, append([]interface{}{level}, args...)...)

	select {
	case l.asyncQueue <- msg:
	default:
		// 如果队列满了，直接写入以避免阻塞
		l.writeLog(msg)
	}
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.logAsync("INFO", format, args...)
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		l.logAsync("DEBUG", format, args...)
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.logAsync("ERROR", format, args...)
}

// Progress 显示进度
func (l *Logger) Progress(current, total int64, startTime time.Time, instantRPS float64, remaining time.Duration) {
	if !l.verbose {
		return
	}

	elapsed := time.Since(startTime)
	percent := float64(current) / float64(total) * 100
	rps := float64(current) / elapsed.Seconds()

	// 构建固定格式的进度信息
	var progressStr string
	if remaining <= 0 {
		progressStr = fmt.Sprintf("\rProgress: %d/%d (%5.1f%%) - %6.1f req/sec - Instant: %6.1f req/sec - Elapsed: %v",
			current, total, percent, rps, instantRPS, elapsed.Round(time.Second))
	} else {
		progressStr = fmt.Sprintf("\rProgress: %d/%d (%5.1f%%) - %6.1f req/sec - Elapsed: %v - Remaining: %v",
			current, total, percent, rps, elapsed.Round(time.Second), remaining.Round(time.Second))
	}

	// 清理行尾并输出
	fmt.Print(progressStr + strings.Repeat(" ", max(0, l.lastLineLength-len(progressStr))))
	l.lastLineLength = len(progressStr)

	// 完成后换行
	if current >= total {
		fmt.Println()
		l.lastLineLength = 0
	}
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	// 停止异步处理
	close(l.asyncStop)
	l.asyncWg.Wait()

	// 停止定期刷新
	if l.flushTicker != nil {
		l.flushTicker.Stop()
		close(l.flushStop)
		l.flushWg.Wait()
	}

	// 最终刷新缓冲区
	l.flush()

	// 关闭文件
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
