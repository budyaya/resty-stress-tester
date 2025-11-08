package types

import (
	"sort"
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

	// 分位数统计 - 新增字段
	P50ResponseTime time.Duration `json:"p50_response_time"`
	P90ResponseTime time.Duration `json:"p90_response_time"`
	P99ResponseTime time.Duration `json:"p99_response_time"`

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

	// 计算分位数 - 新增功能
	sr.calculatePercentiles()
}

// calculatePercentiles 计算响应时间分位数 - 新增方法
func (sr *StressResult) calculatePercentiles() {
	if len(sr.DetailedResults) == 0 {
		return
	}

	// 提取所有成功的响应时间
	var responseTimes []time.Duration
	for _, result := range sr.DetailedResults {
		if result.Success {
			responseTimes = append(responseTimes, result.Duration)
		}
	}

	if len(responseTimes) == 0 {
		return
	}

	// 排序响应时间
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	// 计算分位数
	sr.P50ResponseTime = calculatePercentile(responseTimes, 0.50)
	sr.P90ResponseTime = calculatePercentile(responseTimes, 0.90)
	sr.P99ResponseTime = calculatePercentile(responseTimes, 0.99)
}

// calculatePercentile 计算分位数 - 新增函数
func calculatePercentile(sortedData []time.Duration, percentile float64) time.Duration {
	if len(sortedData) == 0 {
		return 0
	}

	index := percentile * float64(len(sortedData)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sortedData) {
		return sortedData[lower]
	}

	weight := index - float64(lower)
	return time.Duration(float64(sortedData[lower])*(1-weight) + float64(sortedData[upper])*weight)
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
