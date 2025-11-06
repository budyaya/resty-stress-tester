package engine

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// RestyClientConfig RESTy 客户端配置
type RestyClientConfig struct {
	Timeout          time.Duration
	KeepAlive        bool
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
}

// NewRestyClient 创建 RESTy 客户端
func NewRestyClient(cfg *RestyClientConfig) *resty.Client {
	client := resty.New()

	// 基础配置
	client.SetTimeout(cfg.Timeout)

	if !cfg.KeepAlive {
		client.SetCloseConnection(true)
	}

	// TLS 配置
	client.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})

	// 重试配置
	client.SetRetryCount(cfg.RetryCount)
	client.SetRetryWaitTime(cfg.RetryWaitTime)
	client.SetRetryMaxWaitTime(cfg.RetryMaxWaitTime)

	// 调试配置
	if false { // 可根据需要开启
		client.SetDebug(true)
	}

	return client
}

// RequestBuilder 请求构建器
type RequestBuilder struct {
	client *resty.Client
}

// NewRequestBuilder 创建请求构建器
func NewRequestBuilder(client *resty.Client) *RequestBuilder {
	return &RequestBuilder{
		client: client,
	}
}

// BuildRequest 构建请求
func (b *RequestBuilder) BuildRequest(
	method string,
	url string,
	headers map[string]string,
	body interface{},
) *resty.Request {
	req := b.client.R()

	// 设置方法
	req.Method = strings.ToUpper(method)

	// 设置 URL
	req.URL = url

	// 设置 Headers
	if len(headers) > 0 {
		req.SetHeaders(headers)
	}

	// 设置请求体
	if body != nil {
		req.SetBody(body)
	}

	return req
}

// RequestExecutor 请求执行器
type RequestExecutor struct {
	client *resty.Client
}

// NewRequestExecutor 创建请求执行器
func NewRequestExecutor(client *resty.Client) *RequestExecutor {
	return &RequestExecutor{
		client: client,
	}
}

// Execute 执行请求
func (e *RequestExecutor) Execute(req *resty.Request) (*resty.Response, error) {
	return req.Execute(req.Method, req.URL)
}

// ExecuteWithRetry 带重试的执行
func (e *RequestExecutor) ExecuteWithRetry(req *resty.Request, maxRetries int) (*resty.Response, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		resp, err := e.Execute(req)
		if err == nil && resp.StatusCode() < 500 {
			return resp, nil
		}

		lastErr = err
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %v", maxRetries, lastErr)
}
