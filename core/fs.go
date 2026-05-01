package core

import (
	"context"
	"io"
	"time"
)

// Object represents a unified file/directory object across all cloud providers.
type Object struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
	Path    string    `json:"path"`
	Hash    string    `json:"hash,omitempty"`
	RawData interface{} `json:"-"`
}

// UploadPartInfo stores info about a single uploaded chunk.
type UploadPartInfo struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

// UploadState represents the resumable upload checkpoint.
type UploadState struct {
	FilePath      string           `json:"file_path"`
	DestPath      string           `json:"dest_path"`
	TotalSize     int64            `json:"total_size"`
	PartSize      int64            `json:"part_size"`
	UploadedParts []UploadPartInfo `json:"uploaded_parts"`
	UploadID      string           `json:"upload_id"`
	CreatedAt     time.Time        `json:"created_at"`
}

// Storage is the interface that all backend drivers must implement.
type Storage interface {
	// Name returns the driver name.
	Name() string

	// Init initializes the driver with configuration.
	Init(cfg map[string]string) error

	// User returns account information.
	User(ctx context.Context) (map[string]interface{}, error)

	// List returns objects in a directory.
	List(ctx context.Context, pathOrID string) ([]*Object, error)

	// Open opens a remote file for reading (supports Range requests for resuming).
	Open(ctx context.Context, obj *Object, offset int64) (io.ReadCloser, error)

	// Put uploads data to a remote directory (supports streaming via io.Reader).
	Put(ctx context.Context, in io.Reader, remoteDir *Object, name string, size int64) (*Object, error)

	// Delete removes an object.
	Delete(ctx context.Context, obj *Object) error

	// Move moves an object.
	Move(ctx context.Context, src *Object, destDir *Object) error

	// Rename renames an object.
	Rename(ctx context.Context, obj *Object, newName string) error

	// Mkdir creates a directory.
	Mkdir(ctx context.Context, pathOrID string) (*Object, error)

	// Info retrieves detailed metadata for a specific file/folder.
	Info(ctx context.Context, pathOrID string) (*Object, error)

	// Copy copies a file/folder to a destination directory.
	Copy(ctx context.Context, src *Object, destDir *Object) error

	// CreateShare creates a share for the given file IDs.
	CreateShare(ctx context.Context, fileIDs []string, expiredDay int) (*struct{ ShareID, URL string }, error)

	// ListShares lists all shares.
	ListShares(ctx context.Context, page, size int) ([]struct{ ShareID, URL, Title string; IsExpired bool; FileCount int; CreatedAt int64 }, error)

	// DeleteShare deletes one or more shares.
	DeleteShare(ctx context.Context, shareIDs []string) error

	// Search searches for files/folders by query.
	Search(ctx context.Context, query, pdirFid string, page, size int) ([]*Object, error)

	// Quota returns the storage space usage.
	Quota(ctx context.Context) (*struct{ Total, Used int64 }, error)

	// ListRecycle lists items in the recycle bin.
	ListRecycle(ctx context.Context, page, size int) ([]struct{ Fid, FileName string; Size int64; IsDir bool; DeletedAt int64 }, error)

	// RecoverRecycle recovers items from the recycle bin.
	RecoverRecycle(ctx context.Context, fids []string) error

	// DeleteRecycle permanently deletes items from the recycle bin.
	DeleteRecycle(ctx context.Context, fids []string) error

	// Close releases any resources held by the driver.
	Close() error
}
