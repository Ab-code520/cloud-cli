package quark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/Ab-code520/cloud-cli/utils"
)

type Driver struct {
	api   *QuarkAPI
	cache string
}

func NewDriver() core.Storage {
	return &Driver{
		cache: filepath.Join(os.Getenv("HOME"), ".cache", "cloud-cli", "upload_states"),
	}
}

func init() {
	core.Register("quark", NewDriver)
}

func (d *Driver) Name() string { return "quark" }

func (d *Driver) Init(cfg map[string]string) error {
	cookie := cfg["cookie"]
	if cookie == "" {
		return fmt.Errorf("cookie is required")
	}
	d.api = NewAPI(cookie)
	if err := os.MkdirAll(d.cache, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	return nil
}

func (d *Driver) Close() error { return nil }

func (d *Driver) List(ctx context.Context, pathOrID string) ([]*core.Object, error) {
	// Clean up path: "quark:/" -> "/" -> "0"
	if strings.Contains(pathOrID, ":/") {
		parts := strings.SplitN(pathOrID, ":/", 2)
		pathOrID = "/" + parts[1]
	}
	
	pdirFid := pathOrID
	if pdirFid == "" || pdirFid == "/" {
		pdirFid = "0"
	}

	var allObjects []*core.Object
	page := 1
	for {
		req := ListReq{
			PdirFid: pdirFid,
			Page:    page,
			Size:    1000,
			Sort:    "file_type",
			Order:   "asc",
		}
		resp, err := d.api.List(ctx, &req)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.List {
			// Store parent ID in RawData for Delete/Mkdir usage
			meta := map[string]interface{}{"parent_id": pdirFid}
			
		allObjects = append(allObjects, &core.Object{
				ID:      item.Fid,
				Name:    item.FileName,
				Size:    item.Size,
				IsDir:   item.IsDir,
				ModTime: time.Unix(item.UpdatedAt, 0),
				Path:    item.Fid, // Keep Path as ID for compatibility
				Hash:    "",       // List API might not return Hash, need Info for that
				RawData: meta,
			})
		}

		if len(resp.List) < req.Size {
			break
		}
		page++
	}
	return allObjects, nil
}

func (d *Driver) Open(ctx context.Context, obj *core.Object, offset int64) (io.ReadCloser, error) {
	url, err := d.api.GetDownloadURL(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	resp, err := d.api.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("download request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (d *Driver) Put(ctx context.Context, in io.Reader, remoteDir *core.Object, name string, size int64) (*core.Object, error) {
	var fileSHA1 string
	var fileMD5 string
	var fileReader io.ReaderAt

	// Fix #6: Support io.ReaderAt interface instead of strict *os.File
	if ra, ok := in.(io.ReaderAt); ok {
		fileReader = ra
		// If it's a file, rewind and calc SHA1/MD5
		if f, ok := in.(*os.File); ok {
			_, err := f.Seek(0, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("cannot seek file: %w", err)
			}
			// Calculate SHA1
			sha1, err := utils.CalcReaderSHA1(in)
			if err != nil {
				return nil, fmt.Errorf("calc sha1: %w", err)
			}
			fileSHA1 = sha1

			// Rewind and calculate MD5
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("cannot rewind file for MD5: %w", err)
			}
			md5, err := utils.CalcReaderMD5(in)
			if err != nil {
				return nil, fmt.Errorf("calc md5: %w", err)
			}
			fileMD5 = md5

			// Rewind for upload
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("cannot rewind file for upload: %w", err)
			}
		}
	} else {
		return nil, fmt.Errorf("upload requires io.ReaderAt support for concurrent multipart")
	}

	// Dynamic part size: max(10MB, size/9000 + 1MB) to stay within 10000 part limit
	partSize := int64(10 * 1024 * 1024)
	if size > 0 {
		calcSize := size/9000 + 1024*1024
		if calcSize > partSize {
			partSize = calcSize
		}
	}

	preResp, err := d.api.Precreate(ctx, &PrecreateReq{
		PdirFid:    remoteDir.ID,
		FileName:   name,
		Size:       size,
		Sha1:       fileSHA1,
		Md5:        fileMD5,
		ChunkSize:  partSize,
		FormatType: "file",
	})
	if err != nil {
		return nil, fmt.Errorf("precreate failed: %w", err)
	}

	// Use auth_info from precreate response as auth_meta
	authMetaStr := preResp.AuthInfo

	if preResp.RapidUpload {
		return &core.Object{
			ID:   preResp.FileID,
			Name: name,
			Size: size,
		}, nil
	}

	uploadID := preResp.UploadID
	// Use server-provided part size if available, otherwise keep our calculated size
	if preResp.PartSize > 0 && preResp.PartSize > partSize {
		partSize = preResp.PartSize
	}

	totalParts := int((size + partSize - 1) / partSize)

	statePath := d.getStatePath(name, size, uploadID)
	state := d.loadState(statePath)
	if state == nil {
		state = &core.UploadState{
			UploadID:      uploadID,
			TotalSize:     size,
			PartSize:      partSize,
			UploadedParts: []core.UploadPartInfo{},
		}
	}

	auth, err := d.api.Auth(ctx, &AuthReq{TaskID: preResp.TaskID, UploadID: uploadID, AuthMeta: authMetaStr})
	if err != nil {
		return nil, fmt.Errorf("get auth: %w", err)
	}

	pool := utils.NewPool(4)
	var mu sync.Mutex

	// Fix #12: Buffer Pool
	bufPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, partSize)
		},
	}

	uploadedSet := make(map[int]string)
	for _, p := range state.UploadedParts {
		uploadedSet[p.PartNumber] = p.ETag
	}

	for pn := 1; pn <= totalParts; pn++ {
		if ctx.Err() != nil {
			break
		}
		if _, ok := uploadedSet[pn]; ok {
			continue
		}

		partNum := pn
		pool.Go(func(ctx context.Context) error {
			offset := int64(partNum-1) * partSize
			remSize := partSize
			if offset+remSize > size {
				remSize = size - offset
			}

			// Get buffer from pool
			buf := bufPool.Get().([]byte)
			defer bufPool.Put(buf) // Return buffer to pool

			// Read part
			buf = buf[:remSize]
			_, err := fileReader.ReadAt(buf, offset)
			if err != nil && err != io.EOF {
				return fmt.Errorf("read part %d: %w", partNum, err)
			}

			// Fix #4: Local Retry with Exponential Backoff
			var etag string
			var lastErr error
			for attempt := 1; attempt <= 3; attempt++ {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				etag, lastErr = d.api.UploadPart(ctx, auth.Endpoint, buf, partNum, uploadID)
				if lastErr == nil {
					break
				}
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
			
			if lastErr != nil {
				return fmt.Errorf("part %d failed after 3 attempts: %w", partNum, lastErr)
			}

			mu.Lock()
			uploadedSet[partNum] = etag
			state.UploadedParts = append(state.UploadedParts, core.UploadPartInfo{
				PartNumber: partNum,
				ETag:       etag,
			})
			d.saveState(statePath, state)
			mu.Unlock()

			return nil
		})
	}

	if err := pool.Wait(); err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	finishReq := FinishReq{UploadID: uploadID}
	sort.Slice(state.UploadedParts, func(i, j int) bool {
		return state.UploadedParts[i].PartNumber < state.UploadedParts[j].PartNumber
	})

	for _, p := range state.UploadedParts {
		finishReq.Parts = append(finishReq.Parts, struct {
			PartNumber int    `json:"part_number"`
			ETag       string `json:"etag"`
		}{p.PartNumber, p.ETag})
	}

	if err := d.api.FinishUpload(ctx, &finishReq); err != nil {
		return nil, fmt.Errorf("finish upload: %w", err)
	}

	os.Remove(statePath)

	return &core.Object{
		Name: name,
		Size: size,
	}, nil
}

func (d *Driver) User(ctx context.Context) (map[string]interface{}, error) {
	space, err := d.api.GetSpace(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch user info failed: %w", err)
	}
	return map[string]interface{}{
		"type":       "Quark Drive",
		"space_used": space.UsedSpace,
		"space_total": space.TotalCapacity,
	}, nil
}

func (d *Driver) Delete(ctx context.Context, obj *core.Object) error {
	pdirFid := "0"
	if obj.RawData != nil {
		if meta, ok := obj.RawData.(map[string]interface{}); ok {
			if pid, ok := meta["parent_id"].(string); ok {
				pdirFid = pid
			}
		}
	}
	return d.api.DeleteFile(ctx, obj.ID, pdirFid, obj.IsDir)
}

// resolvePathToParent splits a path (e.g., "/a/b/newdir" or "newdir") into (ParentID, DirName).
// It recursively resolves intermediate folders to find the ParentID.
func (d *Driver) resolvePathToParent(ctx context.Context, pathOrID string) (string, string) {
	// Clean path
	pathOrID = strings.Trim(pathOrID, "/")
	if pathOrID == "" {
		return "0", ""
	}

	parts := strings.Split(pathOrID, "/")
	if len(parts) == 1 {
		return "0", parts[0]
	}

	// Resolve parent path
	parentPath := strings.Join(parts[:len(parts)-1], "/")
	dirName := parts[len(parts)-1]

	parentID, err := d.resolvePath(ctx, parentPath)
	if err != nil {
		return "0", dirName // Fallback
	}
	return parentID, dirName
}

// resolvePath converts a relative path (e.g. "folder/subfolder") to a File ID.
func (d *Driver) resolvePath(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "0", nil
	}
	// Strip driver prefix like "quark:"
	if strings.Contains(path, ":") {
		parts := strings.SplitN(path, ":", 2)
		path = parts[1]
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentID := "0"

	for _, part := range parts {
		if part == "" {
			continue
		}
		// List current directory
		objs, err := d.List(ctx, currentID)
		if err != nil {
			return "", err
		}
		found := false
		for _, obj := range objs {
			if obj.Name == part && obj.IsDir {
				currentID = obj.ID
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("directory '%s' not found in '%s'", part, currentID)
		}
	}
	return currentID, nil
}

func (d *Driver) Mkdir(ctx context.Context, pathOrID string) (*core.Object, error) {
	parentID, name := d.resolvePathToParent(ctx, pathOrID)
	if parentID == "" || name == "" {
		return nil, fmt.Errorf("invalid path for mkdir: %s", pathOrID)
	}
	resp, err := d.api.Mkdir(ctx, parentID, name)
	if err != nil {
		return nil, err
	}
	meta := map[string]interface{}{"parent_id": parentID}
	return &core.Object{
		ID:      resp.Fid,
		Name:    name,
		IsDir:   true,
		Path:    pathOrID,
		RawData: meta,
	}, nil
}

func (d *Driver) Info(ctx context.Context, pathOrID string) (*core.Object, error) {
	if pathOrID == "" {
		pathOrID = "0" // Root
	}
	resp, err := d.api.GetFileInfo(ctx, pathOrID)
	if err != nil {
		return nil, err
	}
	return &core.Object{
		ID:      resp.Fid,
		Name:    resp.FileName,
		Size:    resp.Size,
		IsDir:   resp.IsDir,
		ModTime: time.Unix(resp.UpdatedAt, 0),
		Path:    resp.Fid,
		Hash:    resp.Hash,
	}, nil
}

func (d *Driver) Copy(ctx context.Context, src *core.Object, destDir *core.Object) error {
	destID := destDir.ID
	if destID == "" {
		destID = "0"
	}
	return d.api.CopyFile(ctx, src.ID, destID)
}

func (d *Driver) CreateShare(ctx context.Context, fileIDs []string, expiredDay int) (string, error) {
	resp, err := d.api.CreateShare(ctx, fileIDs, expiredDay)
	if err != nil {
		return "", err
	}
	return resp.URL, nil
}

func (d *Driver) ListShares(ctx context.Context, page, size int) ([]core.ShareInfo, error) {
	resp, err := d.api.ListShares(ctx, page, size)
	if err != nil {
		return nil, err
	}
	result := make([]core.ShareInfo, len(resp.List))
	for i, item := range resp.List {
		result[i].ShareID = item.ShareID
		result[i].URL = item.URL
		result[i].Title = item.Title
		result[i].IsExpired = item.IsExpired
		result[i].FileCount = item.FileCount
		result[i].CreatedAt = item.CreatedAt
	}
	return result, nil
}

func (d *Driver) DeleteShare(ctx context.Context, shareIDs []string) error {
	return d.api.DeleteShare(ctx, shareIDs)
}

func (d *Driver) Search(ctx context.Context, query, pdirFid string, page, size int) ([]*core.Object, error) {
	resp, err := d.api.SearchFiles(ctx, query, pdirFid, page, size)
	if err != nil {
		return nil, err
	}
	objects := make([]*core.Object, 0, len(resp.List))
	for _, item := range resp.List {
		objects = append(objects, &core.Object{
			ID:      item.Fid,
			Name:    item.FileName,
			Size:    item.Size,
			IsDir:   item.IsDir,
			ModTime: time.Unix(item.UpdatedAt, 0),
			Path:    item.Fid,
		})
	}
	return objects, nil
}

func (d *Driver) Move(ctx context.Context, src *core.Object, destDir *core.Object) error {
	destID := destDir.ID
	if destID == "" {
		destID = "0"
	}
	return d.api.MoveFile(ctx, src.ID, destID)
}

func (d *Driver) Rename(ctx context.Context, obj *core.Object, newName string) error {
	return d.api.RenameFile(ctx, obj.ID, newName)
}

func (d *Driver) Quota(ctx context.Context) (int64, int64, error) {
	resp, err := d.api.GetSpace(ctx)
	if err != nil {
		return 0, 0, err
	}
	return resp.TotalCapacity, resp.UsedSpace, nil
}

func (d *Driver) ListRecycle(ctx context.Context, page, size int) ([]core.RecycleItem, error) {
	resp, err := d.api.ListRecycle(ctx, page, size)
	if err != nil {
		return nil, err
	}
	result := make([]core.RecycleItem, len(resp.List))
	for i, item := range resp.List {
		result[i].FID = item.Fid
		result[i].FileName = item.FileName
		result[i].Size = item.Size
		result[i].IsDir = item.IsDir
		result[i].DeletedAt = item.DeletedAt
	}
	return result, nil
}

func (d *Driver) RecoverRecycle(ctx context.Context, fids []string) error {
	return d.api.RecoverRecycle(ctx, fids)
}

func (d *Driver) DeleteRecycle(ctx context.Context, fids []string) error {
	return d.api.DeleteRecycle(ctx, fids)
}

func (d *Driver) getStatePath(name string, size int64, uploadID string) string {
	key := fmt.Sprintf("%s_%d_%s", name, size, uploadID)
	hash := fmt.Sprintf("%x", []byte(key))
	return filepath.Join(d.cache, hash+".json")
}

func (d *Driver) saveState(path string, state *core.UploadState) error {
	data, _ := json.Marshal(state)
	return os.WriteFile(path, data, 0644)
}

func (d *Driver) loadState(path string) *core.UploadState {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var s core.UploadState
	if json.Unmarshal(data, &s) != nil {
		return nil
	}
	return &s
}
