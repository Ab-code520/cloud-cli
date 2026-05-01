package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressBar 终端进度条
type ProgressBar struct {
	total    int64
	current  int64
	start    time.Time
	barWidth int
	unit     string
	mu       sync.Mutex
	enabled  bool
}

// NewProgressBar 创建进度条
func NewProgressBar(total int64, unit string, enabled bool) *ProgressBar {
	return &ProgressBar{
		total:    total,
		barWidth: 40,
		unit:     unit,
		start:    time.Now(),
		enabled:  enabled && isTerminal(),
	}
}

func isTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// Update 更新进度
func (pb *ProgressBar) Update(current int64) {
	if !pb.enabled {
		return
	}
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current = current
	pb.render()
}

// Add 增加进度
func (pb *ProgressBar) Add(delta int64) {
	if !pb.enabled {
		return
	}
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current += delta
	pb.render()
}

func (pb *ProgressBar) render() {
	if pb.total <= 0 {
		return
	}

	percent := float64(pb.current) / float64(pb.total)
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(percent * float64(pb.barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.barWidth-filled)

	elapsed := time.Since(pb.start)
	speed := float64(pb.current) / elapsed.Seconds()

	var speedStr string
	switch {
	case speed > 1024*1024*1024:
		speedStr = fmt.Sprintf("%.2f GB/s", speed/1024/1024/1024)
	case speed > 1024*1024:
		speedStr = fmt.Sprintf("%.2f MB/s", speed/1024/1024)
	case speed > 1024:
		speedStr = fmt.Sprintf("%.2f KB/s", speed/1024)
	default:
		speedStr = fmt.Sprintf("%.0f B/s", speed)
	}

	remaining := time.Duration(float64(pb.total-pb.current) / speed * float64(time.Second))
	if remaining < 0 {
		remaining = 0
	}

	fmt.Printf("[%s] %5.1f%% | %s | 剩余: %s",
		bar,
		percent*100,
		speedStr,
		remaining.Truncate(time.Second))

	if percent >= 1.0 {
		fmt.Println()
	}
}

// Done 标记完成
func (pb *ProgressBar) Done() {
	if !pb.enabled {
		return
	}
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current = pb.total
	pb.render()
}

// FormatSize 格式化文件大小
func FormatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024*1024:
		return fmt.Sprintf("%.2f GB", float64(bytes)/1024/1024/1024)
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024/1024)
	case bytes >= 1024:
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatDuration 格式化时间
func FormatDuration(d time.Duration) string {
	d = d.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
