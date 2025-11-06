package types

import (
	"sync"
	"sync/atomic"
	"time"
)

// RequestResult 单个请求结果
type RequestResult struct {
	Timestamp    time.Time     `json:"timestamp"`
	Duration     time.Duration `json:"duration"`
	StatusCode   int           `json:"status_code"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	ResponseSize int           `json:"response_size"`
	CSVData      interface{}   `json:"csv_data,omitempty"`
}

// StressResult 压测结果统计
type StressResult struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	TotalDuration      time.Duration `json:"total_duration"`
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`

	// 响应时间统计
	MinResponseTime   time.Duration `json:"min_response_time"`
	MaxResponseTime   time.Duration `json:"max_response_time"`
	TotalResponseTime int64         `json:"-"` // 用于计算平均值

	// 分布统计
	StatusCodes sync.Map `json:"status_codes"`
	ErrorCounts sync.Map `json:"error_counts"`

	// 详细请求记录
	DetailedResults []*RequestResult `json:"detailed_results,omitempty"`
	mu              sync.RWMutex
}

// NewStressResult 创建新的结果统计器
func NewStressResult() *StressResult {
	return &StressResult{
		StatusCodes:     sync.Map{},
		ErrorCounts:     sync.Map{},
		DetailedResults: make([]*RequestResult, 0),
		MinResponseTime: time.Hour,
	}
}

// AddResult 添加请求结果
func (sr *StressResult) AddResult(result *RequestResult) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	atomic.AddInt64(&sr.TotalRequests, 1)
	atomic.AddInt64(&sr.TotalResponseTime, int64(result.Duration))

	if result.Success {
		atomic.AddInt64(&sr.SuccessfulRequests, 1)

		// 更新状态码统计
		var count int64
		if val, ok := sr.StatusCodes.Load(result.StatusCode); ok {
			count = val.(int64)
		}
		sr.StatusCodes.Store(result.StatusCode, count+1)

		// 更新响应时间统计
		if result.Duration < sr.MinResponseTime {
			sr.MinResponseTime = result.Duration
		}
		if result.Duration > sr.MaxResponseTime {
			sr.MaxResponseTime = result.Duration
		}
	} else {
		atomic.AddInt64(&sr.FailedRequests, 1)

		// 更新错误统计
		var count int64
		if val, ok := sr.ErrorCounts.Load(result.Error); ok {
			count = val.(int64)
		}
		sr.ErrorCounts.Store(result.Error, count+1)
	}

	// 记录详细结果
	if len(sr.DetailedResults) < 10000 { // 限制内存使用
		sr.DetailedResults = append(sr.DetailedResults, result)
	}
}

// CalculateMetrics 计算最终指标
func (sr *StressResult) CalculateMetrics() {
	sr.TotalDuration = sr.EndTime.Sub(sr.StartTime)
}

// ShouldFail 根据错误率决定是否应该失败
func (sr *StressResult) ShouldFail() bool {
	if sr.TotalRequests == 0 {
		return false
	}
	failureRate := float64(sr.FailedRequests) / float64(sr.TotalRequests)
	return failureRate > 0.1 // 10% 错误率阈值
}

// GetRequestsPerSecond 计算每秒请求数
func (sr *StressResult) GetRequestsPerSecond() float64 {
	if sr.TotalDuration == 0 {
		return 0
	}
	return float64(sr.TotalRequests) / sr.TotalDuration.Seconds()
}

// GetAverageResponseTime 计算平均响应时间
func (sr *StressResult) GetAverageResponseTime() time.Duration {
	if sr.TotalRequests == 0 {
		return 0
	}
	return time.Duration(sr.TotalResponseTime / sr.TotalRequests)
}

// GetSuccessRate 计算成功率
func (sr *StressResult) GetSuccessRate() float64 {
	if sr.TotalRequests == 0 {
		return 0
	}
	return float64(sr.SuccessfulRequests) / float64(sr.TotalRequests) * 100
}
