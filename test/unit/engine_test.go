package unit

import (
	"context"
	"testing"
	"time"

	"github.com/budyaya/resty-stress-tester/internal/config"
	"github.com/budyaya/resty-stress-tester/internal/engine"
	"github.com/budyaya/resty-stress-tester/pkg/types"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		StressConfig: &types.StressConfig{
			URL:       "https://httpbin.org/get",
			Method:    "GET",
			Timeout:   10 * time.Second,
			KeepAlive: true,
		},
	}

	// 创建 HTTP 客户端
	client := resty.New()
	client.SetTimeout(cfg.Timeout)

	// 创建结果收集器
	result := types.NewStressResult()

	// 创建工作协程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := engine.NewWorker(cfg, client, nil, nil, result, ctx)

	// 测试请求通道
	requests := make(chan struct{}, 1)
	requests <- struct{}{}
	close(requests)

	// 运行工作协程
	worker.Run(requests)

	// 验证结果
	assert.Greater(t, result.TotalRequests, int64(0))
	assert.Greater(t, result.SuccessfulRequests, int64(0))
}

func TestRequestBuilder(t *testing.T) {
	client := resty.New()
	builder := engine.NewRequestBuilder(client)

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Test":       "value",
	}

	body := map[string]interface{}{
		"id":   123,
		"name": "test",
	}

	req := builder.BuildRequest("POST", "https://httpbin.org/post", headers, body)

	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "https://httpbin.org/post", req.URL)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "value", req.Header.Get("X-Test"))
	assert.NotNil(t, req.Body)
}

func TestRequestExecutor(t *testing.T) {
	client := resty.New()
	executor := engine.NewRequestExecutor(client)
	builder := engine.NewRequestBuilder(client)

	req := builder.BuildRequest("GET", "https://httpbin.org/get", nil, nil)

	resp, err := executor.Execute(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode())
}
