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

// ErrorItem 错误项
type ErrorItem struct {
	Error string
	Count int64
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

	// 分位数统计
	P50ResponseTime time.Duration `json:"p50_response_time"`
	P90ResponseTime time.Duration `json:"p90_response_time"`
	P99ResponseTime time.Duration `json:"p99_response_time"`

	// 分布统计 - 使用更高效的数据结构
	statusCodes     map[int]int64
	errorCounts     map[string]int64
	statusCodesLock sync.RWMutex
	errorCountsLock sync.RWMutex

	// 详细请求记录 - 使用环形缓冲区避免内存无限增长
	DetailedResults []*RequestResult `json:"detailed_results,omitempty"`
	resultsLock     sync.RWMutex
	resultIndex     int
	maxResults      int
}

// NewStressResult 创建新的结果统计器
func NewStressResult() *StressResult {
	return &StressResult{
		statusCodes:     make(map[int]int64),
		errorCounts:     make(map[string]int64),
		DetailedResults: make([]*RequestResult, 0, 1000), // 预分配容量
		MinResponseTime: time.Hour,
		maxResults:      10000, // 限制最大记录数
	}
}

// AddResult 添加请求结果
func (sr *StressResult) AddResult(result *RequestResult) {
	atomic.AddInt64(&sr.TotalRequests, 1)
	atomic.AddInt64(&sr.TotalResponseTime, int64(result.Duration))

	if result.Success {
		atomic.AddInt64(&sr.SuccessfulRequests, 1)

		// 更新状态码统计
		sr.statusCodesLock.Lock()
		sr.statusCodes[result.StatusCode]++
		sr.statusCodesLock.Unlock()

		// 更新响应时间统计
		sr.resultsLock.Lock()
		if result.Duration < sr.MinResponseTime || sr.MinResponseTime == time.Hour {
			sr.MinResponseTime = result.Duration
		}
		if result.Duration > sr.MaxResponseTime {
			sr.MaxResponseTime = result.Duration
		}
		sr.resultsLock.Unlock()
	} else {
		atomic.AddInt64(&sr.FailedRequests, 1)

		// 更新错误统计
		sr.errorCountsLock.Lock()
		sr.errorCounts[result.Error]++
		sr.errorCountsLock.Unlock()
	}

	// 记录详细结果（使用环形缓冲区逻辑）
	sr.resultsLock.Lock()
	defer sr.resultsLock.Unlock()

	if len(sr.DetailedResults) < sr.maxResults {
		// 如果还有空间，直接追加
		sr.DetailedResults = append(sr.DetailedResults, result)
	} else {
		// 使用环形缓冲区覆盖最旧的结果
		sr.DetailedResults[sr.resultIndex] = result
		sr.resultIndex = (sr.resultIndex + 1) % sr.maxResults
	}
}

// GetSortedStatusCodes 获取排序后的状态码列表
func (sr *StressResult) GetSortedStatusCodes() []int {
	sr.statusCodesLock.RLock()
	defer sr.statusCodesLock.RUnlock()

	codes := make([]int, 0, len(sr.statusCodes))
	for code := range sr.statusCodes {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	return codes
}

// GetStatusCodeCount 获取状态码计数
func (sr *StressResult) GetStatusCodeCount(code int) int64 {
	sr.statusCodesLock.RLock()
	defer sr.statusCodesLock.RUnlock()
	return sr.statusCodes[code]
}

// GetSortedErrors 获取排序后的错误列表
func (sr *StressResult) GetSortedErrors() ([]ErrorItem, int64) {
	sr.errorCountsLock.RLock()
	defer sr.errorCountsLock.RUnlock()

	var totalErrors int64
	errorList := make([]ErrorItem, 0, len(sr.errorCounts))

	for errorMsg, count := range sr.errorCounts {
		errorList = append(errorList, ErrorItem{Error: errorMsg, Count: count})
		totalErrors += count
	}

	// 按错误数量排序
	sort.Slice(errorList, func(i, j int) bool {
		return errorList[i].Count > errorList[j].Count
	})

	return errorList, totalErrors
}

// CalculateMetrics 计算最终指标
func (sr *StressResult) CalculateMetrics() {
	sr.TotalDuration = sr.EndTime.Sub(sr.StartTime)

	// 计算分位数
	sr.calculatePercentiles()
}

// calculatePercentiles 计算响应时间分位数
func (sr *StressResult) calculatePercentiles() {
	sr.resultsLock.RLock()
	defer sr.resultsLock.RUnlock()

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

	// 如果数据量很大，使用采样来加速计算
	if len(responseTimes) > 10000 {
		sampled := make([]time.Duration, 10000)
		step := len(responseTimes) / 10000
		for i := 0; i < 10000; i++ {
			sampled[i] = responseTimes[i*step]
		}
		responseTimes = sampled
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

// calculatePercentile 计算分位数
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

// GetMinResponseTime 获取最小响应时间
func (sr *StressResult) GetMinResponseTime() time.Duration {
	sr.resultsLock.RLock()
	defer sr.resultsLock.RUnlock()
	if sr.MinResponseTime == time.Hour {
		return 0
	}
	return sr.MinResponseTime
}

// GetMaxResponseTime 获取最大响应时间
func (sr *StressResult) GetMaxResponseTime() time.Duration {
	sr.resultsLock.RLock()
	defer sr.resultsLock.RUnlock()
	return sr.MaxResponseTime
}

// SetMaxResults 设置最大结果记录数
func (sr *StressResult) SetMaxResults(max int) {
	sr.resultsLock.Lock()
	defer sr.resultsLock.Unlock()
	sr.maxResults = max
	// 如果当前结果数超过新的最大值，截断
	if len(sr.DetailedResults) > max {
		sr.DetailedResults = sr.DetailedResults[:max]
		sr.resultIndex = 0
	}
}
