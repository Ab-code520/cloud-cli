package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

func FormatSize(bytes int64) string {
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func FormatSpeed(bytesPerSec float64) string {
	switch {
	case bytesPerSec >= GB:
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/GB)
	case bytesPerSec >= MB:
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/MB)
	case bytesPerSec >= KB:
		return fmt.Sprintf("%.2f KB/s", bytesPerSec/KB)
	default:
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	}
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

type ProgressBar struct {
	mu          sync.Mutex
	total       int64
	current     int64
	startTime   time.Time
	lastUpdate  time.Time
	lastCurrent int64
	speed       float64
	width       int
}

func NewProgressBar(total int64, width int) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     width,
		startTime: time.Now(),
		lastUpdate: time.Now(),
	}
}

func (pb *ProgressBar) Update(current int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = current
	now := time.Now()
	elapsed := now.Sub(pb.startTime).Seconds()

	if elapsed > 0 {
		deltaT := now.Sub(pb.lastUpdate).Seconds()
		if deltaT > 0 {
			pb.speed = float64(current-pb.lastCurrent) / deltaT
		} else {
			pb.speed = float64(current) / elapsed
		}
	}

	pb.lastUpdate = now
	pb.lastCurrent = current
}

func (pb *ProgressBar) Render() string {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	percent := float64(pb.current) / float64(pb.total) * 100
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(pb.width) * percent / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)

	remaining := time.Duration(float64(pb.total-pb.current)/pb.speed) * time.Second

	return fmt.Sprintf(" %s %6.1f%% %s/s [%s/%s] ETA:%s",
		bar, percent,
		FormatSpeed(pb.speed),
		FormatSize(pb.current),
		FormatSize(pb.total),
		FormatDuration(remaining),
	)
}

// MicroRender outputs a highly compressed progress string for small terminals (e.g., 2.8" screens, mobile SSH).
// Format: "50%|1.2M/s|10s" (fits in ~20 chars)
func (pb *ProgressBar) MicroRender() string {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	percent := float64(pb.current) / float64(pb.total) * 100
	if percent > 100 {
		percent = 100
	}

	remaining := time.Duration(float64(pb.total-pb.current)/pb.speed) * time.Second

	speedStr := FormatSpeed(pb.speed)
	// Compress speed string for micro mode
	if len(speedStr) > 6 {
		speedStr = speedStr[:5]
	}

	return fmt.Sprintf("%3.0f%%|%s|%s", percent, speedStr, FormatDuration(remaining))
}
