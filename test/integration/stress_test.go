package integration

import (
	"os"
	"testing"
	"time"

	"github.com/budyaya/resty-stress-tester/internal/config"
	"github.com/budyaya/resty-stress-tester/internal/engine"
	"github.com/budyaya/resty-stress-tester/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		StressConfig: &types.StressConfig{
			URL:           "https://httpbin.org/get",
			Method:        "GET",
			TotalRequests: 10,
			Concurrency:   2,
			Timeout:       30 * time.Second,
			KeepAlive:     true,
			Verbose:       false,
		},
	}

	tester, err := engine.NewStressEngine(cfg)
	require.NoError(t, err)

	result := tester.Run()

	assert.Equal(t, int64(10), result.TotalRequests)
	assert.Greater(t, result.SuccessfulRequests, int64(8)) // 允许少量失败
	assert.Less(t, result.FailedRequests, int64(3))
	assert.Greater(t, result.GetSuccessRate(), 80.0)
	assert.Greater(t, result.GetRequestsPerSecond(), 0.0)

	tester.Cleanup()
}

func TestCSVParameterization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建临时 CSV 文件
	csvContent := `id,name,category
1,Test User,premium
2,Another User,standard`

	tmpFile, err := os.CreateTemp("", "test*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(csvContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg := &config.Config{
		StressConfig: &types.StressConfig{
			URL:           "https://httpbin.org/anything/user/{{id}}",
			Method:        "GET",
			TotalRequests: 4,
			Concurrency:   2,
			Timeout:       30 * time.Second,
			CSVFile:       tmpFile.Name(),
			Headers: map[string]string{
				"X-User-ID":  "{{id}}",
				"X-Category": "{{category}}",
			},
			Verbose: false,
		},
	}

	tester, err := engine.NewStressEngine(cfg)
	require.NoError(t, err)

	result := tester.Run()

	assert.Equal(t, int64(4), result.TotalRequests)
	assert.Greater(t, result.SuccessfulRequests, int64(2))

	tester.Cleanup()
}

func TestPostRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		StressConfig: &types.StressConfig{
			URL:           "https://httpbin.org/post",
			Method:        "POST",
			TotalRequests: 5,
			Concurrency:   2,
			Timeout:       30 * time.Second,
			Body:          `{"test": "data", "timestamp": "2024-01-01"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Verbose: false,
		},
	}

	tester, err := engine.NewStressEngine(cfg)
	require.NoError(t, err)

	result := tester.Run()

	assert.Equal(t, int64(5), result.TotalRequests)
	assert.Greater(t, result.SuccessfulRequests, int64(3))

	tester.Cleanup()
}
