package core

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// SyncAction represents an action to be taken during sync.
type SyncAction struct {
	Type   string // "upload", "delete", "update", "skip"
	Object *Object
	Reason string
}

// SyncConfig holds the configuration for a sync operation.
type SyncConfig struct {
	// DeleteExtra deletes files in Dest that are not in Source.
	DeleteExtra bool
	// DryRun shows what would be done without executing.
	DryRun bool
	// MaxAge skips files older than this duration.
	MaxAge time.Duration
}

// Syncer performs directory synchronization between two Storage backends.
type Syncer struct {
	Source Storage
	Dest   Storage
	Config SyncConfig
	
	// Callbacks
	OnAction func(action *SyncAction)
}

// Sync performs the synchronization.
func (s *Syncer) Sync(ctx context.Context, srcPath, destPath string) error {
	// 1. List Source
	srcFiles, err := s.Source.List(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("list source: %w", err)
	}

	// 2. List Dest
	destFiles, err := s.Dest.List(ctx, destPath)
	if err != nil {
		// If dest doesn't exist, we might need to create it?
		// Or it's an error.
		return fmt.Errorf("list dest: %w", err)
	}

	// 3. Build maps
	srcMap := make(map[string]*Object)
	for _, f := range srcFiles {
		srcMap[f.Name] = f
	}

	destMap := make(map[string]*Object)
	for _, f := range destFiles {
		destMap[f.Name] = f
	}

	// 4. Compare: Source -> Dest (Upload/Update)
	for _, srcObj := range srcFiles {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		destObj, exists := destMap[srcObj.Name]
		relativeDest := filepath.Join(destPath, srcObj.Name)

		if srcObj.IsDir {
			// Recursively sync directories
			// Ensure dest dir exists
			if !exists {
				if s.Config.DryRun {
					s.reportAction(&SyncAction{Type: "mkdir", Object: srcObj, Reason: "Directory missing in dest"})
				} else {
					_, _ = s.Dest.Mkdir(ctx, relativeDest)
				}
			}
			
			// Sync content
			if err := s.Sync(ctx, filepath.Join(srcPath, srcObj.Name), relativeDest); err != nil {
				return err
			}
			continue
		}

		if !exists {
			// New file
			s.reportAction(&SyncAction{Type: "upload", Object: srcObj, Reason: "New file"})
			if !s.Config.DryRun {
				if err := s.uploadFile(ctx, srcObj, srcPath, destPath); err != nil {
					return fmt.Errorf("upload %s: %w", srcObj.Name, err)
				}
			}
		} else {
			// File exists, check if update needed
			// Simple comparison: Size or ModTime
			// Note: ModTime precision might differ.
			if s.needsUpdate(srcObj, destObj) {
				s.reportAction(&SyncAction{Type: "update", Object: srcObj, Reason: "Modified"})
				if !s.Config.DryRun {
					if err := s.uploadFile(ctx, srcObj, srcPath, destPath); err != nil {
						return fmt.Errorf("update %s: %w", srcObj.Name, err)
					}
				}
			} else {
				s.reportAction(&SyncAction{Type: "skip", Object: srcObj, Reason: "Up to date"})
			}
		}
	}

	// 5. Compare: Dest -> Source (Delete extra)
	if s.Config.DeleteExtra {
		for _, destObj := range destFiles {
			if _, exists := srcMap[destObj.Name]; !exists {
				s.reportAction(&SyncAction{Type: "delete", Object: destObj, Reason: "Extra file in dest"})
				if !s.Config.DryRun {
					if err := s.Dest.Delete(ctx, &Object{
						Path: filepath.Join(destPath, destObj.Name),
						IsDir: destObj.IsDir,
					}); err != nil {
						return fmt.Errorf("delete %s: %w", destObj.Name, err)
					}
				}
			}
		}
	}

	return nil
}

// needsUpdate returns true if srcObj should overwrite destObj.
func (s *Syncer) needsUpdate(src, dest *Object) bool {
	if src.Size != dest.Size {
		return true
	}
	// Allow 2 seconds drift for timestamp inaccuracies
	diff := src.ModTime.Sub(dest.ModTime)
	if diff < 0 {
		diff = -diff
	}
	return diff > 2*time.Second
}

func (s *Syncer) uploadFile(ctx context.Context, srcObj *Object, srcDir, destDir string) error {
	// Open source
	reader, err := s.Source.Open(ctx, srcObj, 0)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Upload to dest
	_, err = s.Dest.Put(ctx, reader, &Object{Path: destDir, IsDir: true}, srcObj.Name, srcObj.Size)
	return err
}

func (s *Syncer) reportAction(action *SyncAction) {
	if s.OnAction != nil {
		s.OnAction(action)
	}
}
