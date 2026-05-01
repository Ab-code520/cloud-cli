package core

import "time"

// Object represents a unified file/directory object across all cloud providers.
type Object struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	IsDir   bool        `json:"is_dir"`
	ModTime time.Time   `json:"mod_time"`
	Path    string      `json:"path"`
	RawData interface{} `json:"-"`
}

// Storage is the interface that all backend drivers must implement.
type Storage interface {
	// Init initializes the driver with configuration (tokens, cookies, etc.)
	Init(cfg map[string]string) error

	// List returns objects in a directory
	List(pathOrID string) ([]*Object, error)

	// Download downloads an object to local path
	Download(obj *Object, localPath string, concurrency int) error

	// Upload uploads a local file to remote directory
	Upload(localPath string, remoteDir *Object, concurrency int) error

	// Delete removes an object
	Delete(obj *Object) error

	// Move moves an object to a new location
	Move(src, dest *Object) error

	// Rename renames an object
	Rename(obj *Object, newName string) error

	// Mkdir creates a directory
	Mkdir(path string) error

	// User returns account information
	User() (map[string]interface{}, error)

	// Name returns the driver name
	Name() string
}
