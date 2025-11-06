package util

import (
	"fmt"
	"strings"
	"time"
)

// Formatter 格式化工具
type Formatter struct{}

// NewFormatter 创建格式化工具
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatDuration 格式化时长
func (f *Formatter) FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return d.String()
	}

	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	}

	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000.0)
	}

	return d.Round(time.Millisecond).String()
}

// FormatBytes 格式化字节大小
func (f *Formatter) FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatNumber 格式化数字
func (f *Formatter) FormatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	s := fmt.Sprintf("%d", n)
	var result strings.Builder

	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	return result.String()
}

// FormatPercentage 格式化百分比
func (f *Formatter) FormatPercentage(value, total float64) string {
	if total == 0 {
		return "0.00%"
	}

	percentage := (value / total) * 100
	return fmt.Sprintf("%.2f%%", percentage)
}

// TruncateString 截断字符串
func (f *Formatter) TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}

// FormatProgressBar 格式化进度条
func (f *Formatter) FormatProgressBar(current, total int64, width int) string {
	if total == 0 {
		return strings.Repeat(" ", width)
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))

	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.1f%%", bar, percentage*100)
}
