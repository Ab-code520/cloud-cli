package quark

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud-cli/core"
)

func init() {
	core.Register("quark", NewDriver)
	core.RegisterInfo("quark", core.DriverInfo{
		Name:        "quark",
		Description: "Quark Cloud Drive (夸克网盘)",
		Author:      "cloud-cli",
	})
}

type QuarkDriver struct {
	api *QuarkAPI
	cfg map[string]string
}

func NewDriver() core.Storage {
	return &QuarkDriver{}
}

func (d *QuarkDriver) Init(cfg map[string]string) error {
	cookie, ok := cfg["cookie"]
	if !ok || cookie == "" {
		return fmt.Errorf("quark cookie is required")
	}
	d.cfg = cfg
	d.api = NewAPI(cookie)
	return nil
}

func (d *QuarkDriver) List(pathOrID string) ([]*core.Object, error) {
	fid := pathOrID
	if !strings.HasPrefix(pathOrID, "fid:") {
		var err error
		fid, err = d.resolvePath(pathOrID)
		if err != nil {
			return nil, fmt.Errorf("resolve path %s: %w", pathOrID, err)
		}
	}

	resp, err := d.api.ListDir(fid, 1, 200)
	if err != nil {
		return nil, err
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data")
	}

	listData, ok := data["list"].([]interface{})
	if !ok {
		return []*core.Object{}, nil
	}

	var objects []*core.Object
	for _, item := range listData {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		obj := &core.Object{RawData: m}
		if fid, ok := m["fid"].(string); ok {
			obj.ID = fid
		}
		if name, ok := m["file_name"].(string); ok {
			obj.Name = name
		}
		if size, ok := m["size"].(float64); ok {
			obj.Size = int64(size)
		}
		if ftype, ok := m["file_type"].(float64); ok {
			obj.IsDir = int(ftype) == 0
		}
		if updatedAt, ok := m["updated_at"].(float64); ok {
			obj.ModTime = time.Unix(int64(updatedAt/1000), 0)
		} else if createdAt, ok := m["created_at"].(float64); ok {
			obj.ModTime = time.Unix(int64(createdAt/1000), 0)
		}
		if path, ok := m["pdir_fid"].(string); ok {
			obj.Path = path
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

func (d *QuarkDriver) resolvePath(path string) (string, error) {
	path = strings.TrimPrefix(path, "/")
	if path == "" || path == "/" {
		return "0", nil
	}

	parts := strings.Split(path, "/")
	currentFid := "0"

	for _, part := range parts {
		if part == "" {
			continue
		}
		resp, err := d.api.ListDir(currentFid, 1, 200)
		if err != nil {
			return "", err
		}

		data, _ := resp["data"].(map[string]interface{})
		listData, _ := data["list"].([]interface{})

		found := false
		for _, item := range listData {
			m, _ := item.(map[string]interface{})
			name, _ := m["file_name"].(string)
			if name == part {
				currentFid, _ = m["fid"].(string)
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("path component '%s' not found", part)
		}
	}

	return currentFid, nil
}

func (d *QuarkDriver) Download(obj *core.Object, localPath string, concurrency int) error {
	if obj.IsDir {
		return fmt.Errorf("cannot download directory")
	}

	resp, err := d.api.GetDownloadURL(obj.ID)
	if err != nil {
		return err
	}

	downloadURL, _ := resp["download_url"].(string)
	if downloadURL == "" {
		return fmt.Errorf("no download URL returned")
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}

	// 下载需要正确的 headers 和 Cookie
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://pan.quark.cn/")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="146", "Google Chrome";v="146", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Dest", "iframe")
	req.Header.Set("Cookie", d.api.cookie)

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("download failed: status %d", httpResp.StatusCode)
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, httpResp.Body)
	return err
}

func (d *QuarkDriver) Upload(localPath string, remoteDir *core.Object, concurrency int) error {
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("directory upload not yet supported")
	}

	fileName := fileInfo.Name()
	fileSize := fileInfo.Size()

	// Calculate MD5 and SHA1
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	md5Hash := md5.New()
	sha1Hash := sha1.New()
	multiWriter := io.MultiWriter(md5Hash, sha1Hash)

	if _, err := io.Copy(multiWriter, f); err != nil {
		return fmt.Errorf("calculate hash: %w", err)
	}

	md5Str := fmt.Sprintf("%x", md5Hash.Sum(nil))
	sha1Str := fmt.Sprintf("%x", sha1Hash.Sum(nil))

	// Get MIME type
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Get target directory fid
	targetFid := remoteDir.ID
	if targetFid == "" {
		targetFid, err = d.resolvePath(remoteDir.Path)
		if err != nil {
			return err
		}
	}

	// Pre-upload
	fmt.Printf("[quark] 预检上传: %s (%s)\n", fileName, formatSize(fileSize))
	preResp, err := d.api.UploadPrepare(fileName, fileSize, targetFid, md5Str, sha1Str, mimeType)
	if err != nil {
		return fmt.Errorf("upload prepare: %w", err)
	}

	// Check 秒传
	if preResp.Data.Finish {
		fmt.Printf("[quark] ✅ 秒传成功: %s\n", fileName)
		return nil
	}

	// Upload parts
	fmt.Printf("[quark] 开始分片上传 (part_size=%s)...\n", formatSize(preResp.Metadata.PartSize))

	f.Seek(0, io.SeekStart)
	partSize := preResp.Metadata.PartSize
	if partSize <= 0 {
		partSize = 4 * 1024 * 1024 // 4MB default
	}

	var etags []string
	partNum := 1

	for {
		chunk := make([]byte, partSize)
		n, err := f.Read(chunk)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			return fmt.Errorf("read chunk: %w", err)
		}

		// Get auth key for this part
		now := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
		authMeta := fmt.Sprintf("PUT\n\n%s\n%s\n", mimeType, now)
		authMeta += fmt.Sprintf("x-oss-date:%s\nx-oss-user-agent:aliyun-sdk-js/1.0.0 Chrome 145.0.0.0 on Windows 10 64-bit\n/%s/%s?partNumber=%d&uploadId=%s",
			now, preResp.Data.Bucket, preResp.Data.ObjKey, partNum, preResp.Data.UploadID)

		authKey, err := d.api.GetAuthKey(preResp.Data.AuthInfo, authMeta, preResp.Data.TaskID)
		if err != nil {
			return fmt.Errorf("get auth key for part %d: %w", partNum, err)
		}

		fmt.Printf("\r[quark] 上传分片 %d...", partNum)

		etag, err := d.api.UploadPart(preResp, mimeType, partNum, chunk[:n], authKey)
		if err != nil {
			return fmt.Errorf("upload part %d: %w", partNum, err)
		}

		etags = append(etags, etag)
		partNum++

		if err == io.EOF {
			break
		}
	}
	fmt.Println()

	// Commit upload
	fmt.Printf("[quark] 提交上传 (%d 个分片)...\n", len(etags))
	if err := d.api.CommitUpload(preResp, etags); err != nil {
		return fmt.Errorf("commit upload: %w", err)
	}

	fmt.Printf("[quark] ✅ 上传完成: %s (%s)\n", fileName, formatSize(fileSize))
	return nil
}

func (d *QuarkDriver) Delete(obj *core.Object) error {
	_, err := d.api.DeleteFile([]string{obj.ID})
	return err
}

func (d *QuarkDriver) Move(src, dest *core.Object) error {
	toFid := dest.ID
	if toFid == "" {
		var err error
		toFid, err = d.resolvePath(dest.Path)
		if err != nil {
			return err
		}
	}
	_, err := d.api.MoveFile([]string{src.ID}, toFid)
	return err
}

func (d *QuarkDriver) Rename(obj *core.Object, newName string) error {
	_, err := d.api.RenameFile(obj.ID, newName)
	return err
}

func (d *QuarkDriver) Mkdir(path string) error {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	currentFid := "0"

	for _, part := range parts {
		if part == "" {
			continue
		}
		// Check if already exists
		resp, err := d.api.ListDir(currentFid, 1, 200)
		if err != nil {
			return err
		}
		data, _ := resp["data"].(map[string]interface{})
		listData, _ := data["list"].([]interface{})

		exists := false
		for _, item := range listData {
			m, _ := item.(map[string]interface{})
			name, _ := m["file_name"].(string)
			ftype, _ := m["file_type"].(float64)
			if name == part && int(ftype) == 0 {
				currentFid, _ = m["fid"].(string)
				exists = true
				break
			}
		}

		if !exists {
			createResp, err := d.api.CreateDir(currentFid, part)
			if err != nil {
				return err
			}
			cdata, _ := createResp["data"].(map[string]interface{})
			if cdata == nil {
				return fmt.Errorf("create dir failed: %v", createResp["message"])
			}
			currentFid, _ = cdata["fid"].(string)
		}
	}

	return nil
}

func (d *QuarkDriver) User() (map[string]interface{}, error) {
	return d.api.User()
}

func formatSize(bytes int64) string {
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
