package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Ab-code520/cloud-cli/core"
)

// Driver implements core.Storage for the local filesystem.
// This allows treating the local disk as a cloud backend, enabling
// seamless sync and transfer between any two Storage implementations.
type Driver struct {
	root string
}

func NewDriver() core.Storage {
	return &Driver{}
}

func init() {
	core.Register("local", NewDriver)
}

func (d *Driver) Name() string { return "local" }

// Init expects a "root" key in cfg to set the base directory.
func (d *Driver) Init(cfg map[string]string) error {
	root := cfg["root"]
	if root == "" {
		// Default to current directory if not specified
		var err error
		root, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	// Clean the root path
	d.root = filepath.Clean(root)
	
	// Ensure root exists
	if err := os.MkdirAll(d.root, 0755); err != nil {
		return fmt.Errorf("failed to create root directory %s: %w", d.root, err)
	}
	return nil
}

func (d *Driver) Close() error { return nil }

func (d *Driver) User(ctx context.Context) (map[string]interface{}, error) {
	hostname, _ := os.Hostname()
	return map[string]interface{}{
		"type":     "Local Filesystem",
		"hostname": hostname,
		"root":     d.root,
	}, nil
}

// resolvePath converts a relative path (from the CLI perspective) to an absolute OS path.
func (d *Driver) resolvePath(relPath string) string {
	// The relPath passed by the CLI is usually relative to the "root" or absolute.
	// For "local" driver used in sync, we often treat the argument as a path relative to root
	// OR if the user provided an absolute path in config, we append to it.
	
	// However, standard behavior for "local" backend in rclone-like tools:
	// root = "/home/user"
	// path = "docs" -> "/home/user/docs"
	
	clean := filepath.Clean(relPath)
	if filepath.IsAbs(clean) {
		// If the provided path is absolute, we check if it's inside root.
		// For flexibility in CLI, we might just use it as is, or strictly enforce root.
		// Let's strictly enforce root for security/consistency.
		if !filepath.HasPrefix(clean, d.root) {
			// In some CLI usage, users might pass absolute paths directly to the driver init,
			// but here let's assume 'root' is the base.
			// If the path passed to List/Open is absolute, it might override root?
			// Let's stick to: path is relative to root.
			// Exception: if path is exactly d.root, use it.
		}
		return filepath.Join(d.root, clean)
	}
	return filepath.Join(d.root, clean)
}

func (d *Driver) List(ctx context.Context, pathOrID string) ([]*core.Object, error) {
	absPath := d.resolvePath(pathOrID)
	
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", absPath, err)
	}

	var objects []*core.Object
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// Skip hidden files? Maybe add an option later.
		
		obj := &core.Object{
			// ID is the full absolute path for local files
			ID:      info.Name(), // Relative to the listed dir
			Name:    info.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
			Path:    filepath.Join(pathOrID, info.Name()), // Relative path for CLI usage
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func (d *Driver) Open(ctx context.Context, obj *core.Object, offset int64) (io.ReadCloser, error) {
	// obj.Path should be relative to root.
	absPath := d.resolvePath(obj.Path)
	
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}

	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			f.Close()
			return nil, err
		}
	}

	return f, nil
}

func (d *Driver) Put(ctx context.Context, in io.Reader, remoteDir *core.Object, name string, size int64) (*core.Object, error) {
	dirPath := d.resolvePath(remoteDir.Path)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dirPath, err)
	}

	destPath := filepath.Join(dirPath, name)
	
	// Check for existing file to support resume/skip? 
	// Basic implementation: create/truncate.
	f, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("create file %s: %w", destPath, err)
	}
	defer f.Close()

	written, err := io.Copy(f, in)
	if err != nil {
		// Partial cleanup?
		os.Remove(destPath)
		return nil, fmt.Errorf("write file %s: %w", destPath, err)
	}
	
	// Update mod time if needed?
	
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &core.Object{
		ID:      info.Name(),
		Name:    name,
		Size:    written, // Use actual written size
		IsDir:   false,
		ModTime: info.ModTime(),
		Path:    filepath.Join(remoteDir.Path, name),
	}, nil
}

func (d *Driver) Delete(ctx context.Context, obj *core.Object) error {
	absPath := d.resolvePath(obj.Path)
	if obj.IsDir {
		return os.RemoveAll(absPath)
	}
	return os.Remove(absPath)
}

func (d *Driver) Move(ctx context.Context, src *core.Object, destDir *core.Object) error {
	srcPath := d.resolvePath(src.Path)
	destPath := filepath.Join(d.resolvePath(destDir.Path), src.Name)
	return os.Rename(srcPath, destPath)
}

func (d *Driver) Rename(ctx context.Context, obj *core.Object, newName string) error {
	srcPath := d.resolvePath(obj.Path)
	destPath := filepath.Join(filepath.Dir(srcPath), newName)
	return os.Rename(srcPath, destPath)
}

func (d *Driver) Mkdir(ctx context.Context, pathOrID string) (*core.Object, error) {
	absPath := d.resolvePath(pathOrID)
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, err
	}
	info, _ := os.Stat(absPath)
	return &core.Object{
		ID:      filepath.Base(absPath),
		Name:    filepath.Base(absPath),
		IsDir:   true,
		ModTime: info.ModTime(),
		Path:    pathOrID,
	}, nil
}

func (d *Driver) Info(ctx context.Context, pathOrID string) (*core.Object, error) {
	absPath := d.resolvePath(pathOrID)
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	return &core.Object{
		ID:      info.Name(),
		Name:    info.Name(),
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
		Path:    pathOrID,
	}, nil
}

func (d *Driver) Copy(ctx context.Context, src *core.Object, destDir *core.Object) error {
	// Local copy is tricky if we just use cp.
	// But standard "Copy" in cloud usually means server-side copy.
	// For local, we can implement file copying, but let's skip for now 
	// or just use the Transfer engine which uses Open->Put.
	return fmt.Errorf("local copy (server-side) not supported, use transfer/sync")
}

// LocalFS doesn't support cloud-specific features.
