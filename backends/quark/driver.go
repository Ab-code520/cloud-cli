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
	req := ListReq{
		PdirFid: pathOrID,
		Page:    1,
		Size:    100,
		Sort:    "file_type",
		Order:   "asc",
	}
	if pathOrID == "" {
		req.PdirFid = "0"
	}

	resp, err := d.api.List(ctx, &req)
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
	var seekableFile *os.File

	if f, ok := in.(*os.File); ok {
		_, err := f.Seek(0, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("cannot seek file: %w", err)
		}
		sha1, err := utils.CalcReaderSHA1(in)
		if err != nil {
			return nil, fmt.Errorf("calc sha1: %w", err)
		}
		fileSHA1 = sha1
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("cannot rewind file: %w", err)
		}
		seekableFile = f
	}

	partSize := int64(10 * 1024 * 1024)
	preResp, err := d.api.Precreate(ctx, &PrecreateReq{
		PdirFid:   remoteDir.ID,
		FileName:  name,
		Size:      size,
		Sha1:      fileSHA1,
		ChunkSize: partSize,
	})
	if err != nil {
		return nil, fmt.Errorf("precreate failed: %w", err)
	}

	if preResp.RapidUpload {
		return &core.Object{
			ID:   preResp.FileID,
			Name: name,
			Size: size,
		}, nil
	}

	uploadID := preResp.UploadID
	partSize = preResp.PartSize
	if partSize <= 0 {
		partSize = 10 * 1024 * 1024
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

	auth, err := d.api.GetUploadAuth(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("get auth: %w", err)
	}

	pool := utils.NewPool(4)
	var mu sync.Mutex

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

			buf := make([]byte, remSize)
			if seekableFile != nil {
				_, err := seekableFile.ReadAt(buf, offset)
				if err != nil && err != io.EOF {
					return fmt.Errorf("read part %d: %w", partNum, err)
				}
			} else {
				return fmt.Errorf("concurrent upload requires seekable input")
			}

			etag, err := d.api.UploadPart(ctx, auth.Endpoint, buf, partNum, uploadID)
			if err != nil {
				return err
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
		finishReq.Parts = append(finishReq.Parts, struct{
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

func (d *Driver) User(ctx context.Context) (map[string]interface{}, error) { return nil, nil }
func (d *Driver) Delete(ctx context.Context, obj *core.Object) error      { return nil }
func (d *Driver) Move(ctx context.Context, src *core.Object, destDir *core.Object) error {
	return nil
}
func (d *Driver) Rename(ctx context.Context, obj *core.Object, newName string) error { return nil }
func (d *Driver) Mkdir(ctx context.Context, pathOrID string) (*core.Object, error)   { return nil, nil }

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
