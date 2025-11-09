package engine

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/budyaya/resty-stress-tester/internal/config"
	"github.com/budyaya/resty-stress-tester/internal/parser"
	"github.com/budyaya/resty-stress-tester/pkg/types"
	"github.com/go-resty/resty/v2"
)

// Worker 工作协程
type Worker struct {
	config     *config.Config
	client     *resty.Client
	csvParser  *parser.CSVParser
	tmplParser *parser.TemplateParser
	result     *types.StressResult
	ctx        context.Context
	requestID  int64
	// 复用请求对象减少分配
	baseRequest *resty.Request
}

// NewWorker 创建工作协程
func NewWorker(
	cfg *config.Config,
	client *resty.Client,
	csvParser *parser.CSVParser,
	tmplParser *parser.TemplateParser,
	result *types.StressResult,
	ctx context.Context,
) *Worker {
	worker := &Worker{
		config:     cfg,
		client:     client,
		csvParser:  csvParser,
		tmplParser: tmplParser,
		result:     result,
		ctx:        ctx,
	}

	// 预创建基础请求对象
	worker.baseRequest = client.R().SetContext(ctx)

	return worker
}

// Run 运行工作协程
func (w *Worker) Run(requests <-chan struct{}) {
	for {
		select {
		case <-w.ctx.Done():
			return
		case _, ok := <-requests:
			if !ok {
				return
			}
			w.makeRequest()
		}
	}
}

// makeRequest 发送单个请求
func (w *Worker) makeRequest() {
	startTime := time.Now()

	// 获取 CSV 数据
	var csvData map[string]string
	if w.csvParser != nil {
		requestID := atomic.AddInt64(&w.requestID, 1)
		csvData = w.csvParser.GetRow(int(requestID - 1))
	}

	// 复用基础请求对象
	req := w.baseRequest

	// 处理 URL
	url := w.tmplParser.ProcessURL(w.config.URL, csvData)

	// 处理 Headers
	if len(w.config.Headers) > 0 {
		headers := w.tmplParser.ProcessHeaders(w.config.Headers, csvData)
		req.SetHeaders(headers)
	} else {
		// 清除可能存在的headers
		req.Header = make(map[string][]string)
	}

	// 处理请求体
	if w.config.Body != "" {
		body, err := w.tmplParser.ProcessJSON(w.config.Body, csvData)
		if err != nil {
			w.recordError(startTime, fmt.Sprintf("Failed to process body template: %v", err), csvData)
			return
		}
		req.SetBody(body)
	} else {
		req.SetBody(nil)
	}

	// 发送请求
	var resp *resty.Response
	var err error

	switch strings.ToUpper(w.config.Method) {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	case "PATCH":
		resp, err = req.Patch(url)
	case "HEAD":
		resp, err = req.Head(url)
	case "OPTIONS":
		resp, err = req.Execute("OPTIONS", url)
	default:
		err = fmt.Errorf("unsupported HTTP method: %s", w.config.Method)
	}

	duration := time.Since(startTime)
	w.recordResult(resp, err, duration, csvData)
}

// recordResult 记录请求结果
func (w *Worker) recordResult(resp *resty.Response, err error, duration time.Duration, csvData map[string]string) {
	result := &types.RequestResult{
		Timestamp: time.Now(),
		Duration:  duration,
		CSVData:   csvData,
	}

	if err != nil {
		result.Success = false
		result.Error = w.sanitizeError(err)
	} else {
		result.Success = true
		result.StatusCode = resp.StatusCode()
		result.ResponseSize = len(resp.Body())

		// 检查 HTTP 错误状态码
		if resp.StatusCode() >= 400 {
			result.Success = false
			// 对于HTTP错误，提供更详细的错误信息
			if len(resp.Body()) > 0 {
				// 截断过长的响应体
				body := string(resp.Body())
				if len(body) > 200 {
					body = body[:200] + "..."
				}
				result.Error = fmt.Sprintf("HTTP %d: %s - %s", resp.StatusCode(), resp.Status(), body)
			} else {
				result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.Status())
			}
		}
	}

	w.result.AddResult(result)
}

// recordError 记录错误
func (w *Worker) recordError(startTime time.Time, errorMsg string, csvData map[string]string) {
	result := &types.RequestResult{
		Timestamp: time.Now(),
		Duration:  time.Since(startTime),
		Success:   false,
		Error:     errorMsg,
		CSVData:   csvData,
	}

	w.result.AddResult(result)
}

// sanitizeError 清理错误信息
func (w *Worker) sanitizeError(err error) string {
	if err == nil {
		return ""
	}

	errorMsg := err.Error()

	// 截断过长的错误信息
	if len(errorMsg) > 200 {
		errorMsg = errorMsg[:197] + "..."
	}

	return errorMsg
}
