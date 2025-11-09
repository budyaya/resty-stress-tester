package engine

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/budyaya/resty-stress-tester/internal/config"
	"github.com/budyaya/resty-stress-tester/internal/parser"
	"github.com/budyaya/resty-stress-tester/internal/reporter"
	"github.com/budyaya/resty-stress-tester/internal/util"
	"github.com/budyaya/resty-stress-tester/pkg/types"
	"github.com/go-resty/resty/v2"
)

// StressEngine 压测引擎
type StressEngine struct {
	config     *config.Config
	client     *resty.Client
	csvParser  *parser.CSVParser
	tmplParser *parser.TemplateParser
	reporter   *reporter.StressReporter
	logger     *util.Logger
	result     *types.StressResult
	workers    []*Worker
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	startTime  time.Time
	stopped    int32
}

// NewStressEngine 创建压测引擎
func NewStressEngine(cfg *config.Config) (*StressEngine, error) {
	// 创建 HTTP 客户端
	client := resty.New()
	client.SetTimeout(cfg.Timeout)

	if !cfg.KeepAlive {
		client.SetCloseConnection(true)
	}

	// 配置 TLS
	client.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})

	// 设置重试策略
	client.SetRetryCount(0)

	// 优化连接池
	client.SetTransport(&http.Transport{
		MaxIdleConns:        cfg.Concurrency * 2,
		MaxIdleConnsPerHost: cfg.Concurrency,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   !cfg.KeepAlive,
	})

	// 创建 CSV 解析器
	var csvParser *parser.CSVParser
	if cfg.CSVFile != "" {
		var err error
		csvParser, err = parser.NewCSVParser(cfg.CSVFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create CSV parser: %v", err)
		}
	}

	// 创建模板解析器
	tmplParser := parser.NewTemplateParser(csvParser)

	// 创建日志记录器
	logger, err := util.NewLogger(cfg.Verbose, cfg.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	// 创建报告生成器
	reporter := reporter.NewReporter(cfg)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	return &StressEngine{
		config:     cfg,
		client:     client,
		csvParser:  csvParser,
		tmplParser: tmplParser,
		reporter:   reporter,
		logger:     logger,
		result:     types.NewStressResult(),
		ctx:        ctx,
		cancel:     cancel,
		workers:    make([]*Worker, 0, cfg.Concurrency),
	}, nil
}

// Run 运行压测
func (e *StressEngine) Run() *types.StressResult {
	e.logger.Info("Starting stress test...")
	e.logger.Info("URL: %s", e.config.URL)
	e.logger.Info("Method: %s", e.config.Method)
	e.logger.Info("Concurrency: %d", e.config.Concurrency)

	if e.config.IsDurationBased() {
		e.logger.Info("Duration: %v", e.config.Duration)
	} else {
		e.logger.Info("Total Requests: %d", e.config.TotalRequests)
	}

	if e.csvParser != nil {
		e.logger.Info("CSV Data Rows: %d", e.csvParser.RowCount())
	}

	e.startTime = time.Now()
	e.result.StartTime = e.startTime

	// 预热工作协程
	e.startWorkers()

	// 启动进度监控
	if e.config.Verbose {
		go e.monitorProgress()
	}

	// 等待测试完成
	e.waitForCompletion()

	e.result.EndTime = time.Now()
	e.result.CalculateMetrics()

	e.logger.Info("Stress test completed")

	return e.result
}

// startWorkers 启动工作协程
func (e *StressEngine) startWorkers() {
	// 使用缓冲channel提高性能
	requests := make(chan struct{}, e.config.Concurrency*2)

	// 预创建工作协程
	for i := 0; i < e.config.Concurrency; i++ {
		worker := NewWorker(e.config, e.client, e.csvParser, e.tmplParser, e.result, e.ctx)
		e.workers = append(e.workers, worker)

		e.wg.Add(1)
		go func(w *Worker) {
			defer e.wg.Done()
			w.Run(requests)
		}(worker)
	}

	// 发送请求任务
	go e.sendRequests(requests)
}

// sendRequests 发送请求任务
func (e *StressEngine) sendRequests(requests chan<- struct{}) {
	defer close(requests)

	if e.config.IsDurationBased() {
		// 基于时间的测试
		timer := time.NewTimer(e.config.Duration)
		defer timer.Stop()

		batchSize := e.config.Concurrency
		batch := make([]struct{}, batchSize)

		for {
			select {
			case <-timer.C:
				return
			case <-e.ctx.Done():
				return
			default:
				// 批量发送请求，减少channel操作
				for i := 0; i < batchSize; i++ {
					select {
					case requests <- batch[i]:
					case <-e.ctx.Done():
						return
					default:
						// 如果channel满了，短暂等待
						time.Sleep(10 * time.Microsecond)
					}
				}
			}
		}
	} else {
		// 基于请求数量的测试 - 使用批量发送
		batchSize := min(100, e.config.Concurrency)
		remaining := e.config.TotalRequests

		for remaining > 0 {
			currentBatch := min(batchSize, remaining)
			for i := 0; i < currentBatch; i++ {
				select {
				case requests <- struct{}{}:
				case <-e.ctx.Done():
					return
				}
			}
			remaining -= currentBatch
		}
	}
}

// waitForCompletion 等待测试完成
func (e *StressEngine) waitForCompletion() {
	e.wg.Wait()
}

// monitorProgress 监控进度
func (e *StressEngine) monitorProgress() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastCount int64
	var lastTime time.Time

	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&e.stopped) == 1 {
				return
			}

			current := atomic.LoadInt64(&e.result.TotalRequests)
			now := time.Now()
			elapsed := now.Sub(e.startTime)

			if e.config.IsDurationBased() {
				remaining := e.config.Duration - elapsed
				rps := float64(current) / elapsed.Seconds()
				e.logger.Progress(current, 0, e.startTime)
				e.logger.Debug(" - Elapsed: %v, Remaining: %v, RPS: %.1f",
					elapsed.Round(time.Second), remaining.Round(time.Second), rps)
			} else {
				total := int64(e.config.TotalRequests)
				e.logger.Progress(current, total, e.startTime)

				// 计算瞬时RPS
				if !lastTime.IsZero() {
					instantRPS := float64(current-lastCount) / now.Sub(lastTime).Seconds()
					e.logger.Debug("Instant RPS: %.1f", instantRPS)
				}
				lastCount = current
				lastTime = now
			}

		case <-e.ctx.Done():
			return
		}
	}
}

// Stop 停止压测
func (e *StressEngine) Stop() {
	if atomic.CompareAndSwapInt32(&e.stopped, 0, 1) {
		e.logger.Info("Stopping stress test...")
		if e.cancel != nil {
			e.cancel()
		}
	}
}

// GenerateReport 生成报告
func (e *StressEngine) GenerateReport() error {
	return e.reporter.GenerateReport(e.result)
}

// PrintReport 打印报告
func (e *StressEngine) PrintReport() {
	e.reporter.ConsoleReport(e.result)
}

// Cleanup 清理资源
func (e *StressEngine) Cleanup() {
	e.Stop()
	e.logger.Close()
	if e.client != nil {
		e.client.GetClient().CloseIdleConnections()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
