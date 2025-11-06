package version

import "fmt"

// 版本信息
var (
	Version   = "v0.1.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = "unknown"
)

// String 返回版本信息字符串
func String() string {
	return fmt.Sprintf("Resty-Stress-Tester %s (Build: %s, Commit: %s, Go: %s)",
		Version, BuildTime, GitCommit, GoVersion)
}
