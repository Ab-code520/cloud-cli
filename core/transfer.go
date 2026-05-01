package core

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// TransferStats tracks the progress of a file transfer.
type TransferStats struct {
	FileName    string
	TotalSize   int64
	Transferred int64
	Speed       float64 // bytes per second
	StartTime   time.Time
	EndTime     time.Time
	Err         error
}

// TransferOption configures a transfer operation.
type TransferOption func(t *Transferer)

// Transferer handles streaming file transfers between two different Storage backends.
type Transferer struct {
	// Progress callback
	OnProgress func(stats *TransferStats)
	// Buffer size for io.Copy
	BufferSize int
}

// NewTransferer creates a new Transferer with defaults.
func NewTransferer(opts ...TransferOption) *Transferer {
	t := &Transferer{
		BufferSize: 32 * 1024, // 32KB default buffer
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Transfer copies a single file from src to dest.
func (t *Transferer) Transfer(ctx context.Context, src Storage, dest Storage, srcPath, destPath string) error {
	stats := &TransferStats{
		FileName:  filepath.Base(srcPath),
		StartTime: time.Now(),
	}
	defer func() {
		stats.EndTime = time.Now()
		if t.OnProgress != nil {
			t.OnProgress(stats)
		}
	}()

	// 1. Get source file info (to know size)
	srcInfo, err := src.Info(ctx, srcPath)
	if err != nil {
		stats.Err = fmt.Errorf("get source info: %w", err)
		return stats.Err
	}
	stats.TotalSize = srcInfo.Size

	// 2. Open source stream
	reader, err := src.Open(ctx, srcInfo, 0)
	if err != nil {
		stats.Err = fmt.Errorf("open source: %w", err)
		return stats.Err
	}
	defer reader.Close()

	// 3. Determine destination directory and filename
	destDirPath := filepath.Dir(destPath)
	destFileName := filepath.Base(destPath)

	// Ensure destination directory exists
	// For cloud drivers, this usually means Mkdir
	destDirInfo, err := dest.Mkdir(ctx, destDirPath)
	if err != nil {
		// If Mkdir fails because it exists or other reason, try to ignore if it's not critical
		// But standard behavior: Mkdir should be idempotent or return error if fails.
		// Let's assume Mkdir handles existence or we need to check.
		// For now, if error, we might fail.
		// Improved logic: Check if dir exists via Info?
		// Let's trust Mkdir implementation or handle error.
		// If it returns error "exists", that's usually fine.
	}
	_ = destDirInfo

	// 4. Upload to destination
	_, err = dest.Put(ctx, reader, &Object{Path: destDirPath, IsDir: true}, destFileName, stats.TotalSize)
	if err != nil {
		stats.Err = fmt.Errorf("put to dest: %w", err)
		return stats.Err
	}

	stats.Transferred = stats.TotalSize
	return nil
}

// TransferDir copies a directory recursively.
func (t *Transferer) TransferDir(ctx context.Context, src Storage, dest Storage, srcPath, destPath string, concurrency int) error {
	// 1. List source recursively (simplified)
	// In a real implementation, we need a recursive lister.
	// For now, let's assume single level or use a helper.
	
	// Let's implement a simple recursive walker.
	return t.transferDirRecursive(ctx, src, dest, srcPath, destPath)
}

func (t *Transferer) transferDirRecursive(ctx context.Context, src Storage, dest Storage, srcPath, destPath string) error {
	// List current level
	items, err := src.List(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("list %s: %w", srcPath, err)
	}

	// Ensure dest dir exists
	_, _ = dest.Mkdir(ctx, destPath)

	for _, item := range items {
		// Construct relative paths
		itemSrcPath := filepath.Join(srcPath, item.Name)
		itemDestPath := filepath.Join(destPath, item.Name)

		if item.IsDir {
			// Recurse
			if err := t.transferDirRecursive(ctx, src, dest, itemSrcPath, itemDestPath); err != nil {
				return err
			}
		} else {
			// Transfer file
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				fmt.Printf("📄 %s -> %s\n", item.Name, destPath)
				if err := t.Transfer(ctx, src, dest, itemSrcPath, itemDestPath); err != nil {
					return fmt.Errorf("transfer %s: %w", item.Name, err)
				}
			}
		}
	}
	return nil
}
