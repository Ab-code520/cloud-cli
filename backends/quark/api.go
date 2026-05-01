package quark

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL       = "https://drive-pc.quark.cn"
	apiVersion    = "/1/clouddrive"
	uploadPrePath = "/1/clouddrive/file/upload/pre"
	uploadAuthPath = "/1/clouddrive/file/upload/auth"
	uploadFinishPath = "/1/clouddrive/file/upload/finish"
	downloadPath  = "/1/clouddrive/file/download"
)

type QuarkAPI struct {
	client   *http.Client
	cookie   string
	cookies  map[string]string
}

func NewAPI(cookie string) *QuarkAPI {
	cookies := make(map[string]string)
	for _, pair := range strings.Split(cookie, ";") {
		pair = strings.TrimSpace(pair)
		if idx := strings.Index(pair, "="); idx > 0 {
			cookies[pair[:idx]] = pair[idx+1:]
		}
	}
	return &QuarkAPI{
		client:  &http.Client{},
		cookie:  cookie,
		cookies: cookies,
	}
}

func (a *QuarkAPI) doJSONRequest(method, urlStr string, payload interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	a.setAPIHeaders(req)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		raw := string(respBody)
		if len(raw) > 300 {
			raw = raw[:300]
		}
		return nil, fmt.Errorf("parse response: %w, raw: %s", err, raw)
	}

	return result, nil
}

func (a *QuarkAPI) setAPIHeaders(req *http.Request) {
	req.Header.Set("Cookie", a.cookie)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://pan.quark.cn")
	req.Header.Set("Referer", "https://pan.quark.cn/list")
}

// PreUploadResponse 预上传响应
type PreUploadResponse struct {
	Code   int    `json:"code"`
	Status int    `json:"status"`
	Data   struct {
		Finish  bool   `json:"finish"`
		Fid     string `json:"fid"`
		UploadID string `json:"upload_id"`
		TaskID  string `json:"task_id"`
		Bucket  string `json:"bucket"`
		ObjKey  string `json:"obj_key"`
		UploadURL string `json:"upload_url"`
		AuthInfo json.RawMessage `json:"auth_info"`
		Callback json.RawMessage `json:"callback"`
	} `json:"data"`
	Metadata struct {
		PartSize   int64 `json:"part_size"`
		PartThread int   `json:"part_thread"`
	} `json:"metadata"`
}

// UploadPrepare 上传预检
func (a *QuarkAPI) UploadPrepare(fileName string, fileSize int64, pdirFid string, md5Str string, sha1Str string, mimeType string) (*PreUploadResponse, error) {
	payload := map[string]interface{}{
		"pdir_fid":    pdirFid,
		"file_name":   fileName,
		"size":        fileSize,
		"md5":         md5Str,
		"sha1":        sha1Str,
		"format_type": mimeType,
	}
	u := fmt.Sprintf("%s%s?pr=ucpro&fr=pc", baseURL, uploadPrePath)
	resp, err := a.doJSONRequest("POST", u, payload)
	if err != nil {
		return nil, err
	}

	var preResp PreUploadResponse
	data, _ := json.Marshal(resp)
	if err := json.Unmarshal(data, &preResp); err != nil {
		return nil, fmt.Errorf("parse pre-upload response: %w", err)
	}

	if preResp.Code != 0 || preResp.Status != 200 {
		return &preResp, fmt.Errorf("pre-upload failed: code=%d, status=%d", preResp.Code, preResp.Status)
	}

	return &preResp, nil
}

// GetAuthKey 获取 OSS 请求的 Authorization Key
func (a *QuarkAPI) GetAuthKey(authInfo json.RawMessage, authMeta string, taskID string) (string, error) {
	payload := map[string]interface{}{
		"auth_info": authInfo,
		"auth_meta": authMeta,
		"task_id":   taskID,
	}
	u := fmt.Sprintf("%s%s?pr=ucpro&fr=pc", baseURL, uploadAuthPath)
	resp, err := a.doJSONRequest("POST", u, payload)
	if err != nil {
		return "", err
	}

	code, _ := resp["code"].(float64)
	data, _ := resp["data"].(map[string]interface{})
	if data == nil || code != 0 {
		return "", fmt.Errorf("auth failed: code=%.0f, message=%v", code, resp["message"])
	}

	authKey, _ := data["auth_key"].(string)
	return authKey, nil
}

// UploadPart 上传分片到 OSS
func (a *QuarkAPI) UploadPart(pre *PreUploadResponse, mimeType string, partNumber int, chunkData []byte, authKey string) (string, error) {
	now := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	uploadURLBase := pre.Data.UploadURL
	uploadURLBase = strings.TrimPrefix(uploadURLBase, "https://")
	uploadURLBase = strings.TrimPrefix(uploadURLBase, "http://")

	ossURL := fmt.Sprintf("https://%s.%s/%s?partNumber=%d&uploadId=%s",
		pre.Data.Bucket, uploadURLBase, pre.Data.ObjKey, partNumber, pre.Data.UploadID)

	req, err := http.NewRequest("PUT", ossURL, bytes.NewReader(chunkData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", authKey)
	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("x-oss-date", now)
	req.Header.Set("x-oss-user-agent", "aliyun-sdk-js/1.0.0 Chrome 145.0.0.0 on Windows 10 64-bit")

	// 使用普通 HTTP 客户端
	ossClient := &http.Client{}

	resp, err := ossClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload part %d: %w", partNumber, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("upload part %d failed: status=%d", partNumber, resp.StatusCode)
	}

	etag := resp.Header.Get("ETag")
	return strings.Trim(etag, "\""), nil
}

// CommitUpload 完成分片上传（向 OSS 发送 CompleteMultipartUpload）
func (a *QuarkAPI) CommitUpload(pre *PreUploadResponse, etags []string) error {
	// 构建 XML
	xmlParts := make([]string, len(etags))
	for i, etag := range etags {
		xmlParts[i] = fmt.Sprintf("<Part><PartNumber>%d</PartNumber><ETag>%s</ETag></Part>", i+1, etag)
	}
	xmlBody := fmt.Sprintf("<CompleteMultipartUpload>%s</CompleteMultipartUpload>", strings.Join(xmlParts, ""))

	// Content-MD5
	hash := md5.Sum([]byte(xmlBody))
	contentMD5 := base64.StdEncoding.EncodeToString(hash[:])

	// Callback base64
	var callbackB64 string
	var callbackObj map[string]interface{}
	if err := json.Unmarshal(pre.Data.Callback, &callbackObj); err == nil {
		callbackJSON, _ := json.Marshal(callbackObj)
		callbackB64 = base64.StdEncoding.EncodeToString(callbackJSON)
	} else {
		callbackB64 = base64.StdEncoding.EncodeToString(pre.Data.Callback)
	}

	now := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	// Auth meta for commit
	authMeta := fmt.Sprintf("POST\n%s\napplication/xml\n%s\nx-oss-callback:%s\nx-oss-date:%s\nx-oss-user-agent:aliyun-sdk-js/1.0.0 Chrome 145.0.0.0 on Windows 10 64-bit\n/%s/%s?uploadId=%s",
		contentMD5, now, callbackB64, now, pre.Data.Bucket, pre.Data.ObjKey, pre.Data.UploadID)

	// Get auth key
	authKey, err := a.GetAuthKey(pre.Data.AuthInfo, authMeta, pre.Data.TaskID)
	if err != nil {
		return fmt.Errorf("get commit auth key: %w", err)
	}

	// Build commit URL
	uploadURLBase := strings.TrimPrefix(pre.Data.UploadURL, "https://")
	uploadURLBase = strings.TrimPrefix(uploadURLBase, "http://")
	commitURL := fmt.Sprintf("https://%s.%s/%s?uploadId=%s",
		pre.Data.Bucket, uploadURLBase, pre.Data.ObjKey, pre.Data.UploadID)

	req, err := http.NewRequest("POST", commitURL, strings.NewReader(xmlBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", authKey)
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Content-MD5", contentMD5)
	req.Header.Set("x-oss-callback", callbackB64)
	req.Header.Set("x-oss-date", now)
	req.Header.Set("x-oss-user-agent", "aliyun-sdk-js/1.0.0 Chrome 145.0.0.0 on Windows 10 64-bit")

	ossClient := &http.Client{}

	resp, err := ossClient.Do(req)
	if err != nil {
		return fmt.Errorf("commit upload: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("commit upload failed: status=%d", resp.StatusCode)
	}

	// Call finish API to finalize
	finishPayload := map[string]interface{}{
		"upload_id":    pre.Data.UploadID,
		"fid":          pre.Data.Fid,
		"task_id":      pre.Data.TaskID,
		"uc_param_str": "",
	}
	u := fmt.Sprintf("%s%s?pr=ucpro&fr=pc", baseURL, uploadFinishPath)
	finishResp, err := a.doJSONRequest("POST", u, finishPayload)
	if err != nil {
		return fmt.Errorf("finish request: %w", err)
	}

	fCode, _ := finishResp["code"].(float64)
	if fCode != 0 {
		return fmt.Errorf("finish failed: code=%.0f, msg=%v", fCode, finishResp["message"])
	}

	return nil
}

// GetDownloadURL 获取下载链接
func (a *QuarkAPI) GetDownloadURL(fid string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"fids": []string{fid},
	}
	u := fmt.Sprintf("%s%s?pr=ucpro&fr=pc", baseURL, downloadPath)
	resp, err := a.doJSONRequest("POST", u, payload)
	if err != nil {
		return nil, err
	}

	data, _ := resp["data"].([]interface{})
	if len(data) == 0 {
		return nil, fmt.Errorf("no download data returned")
	}

	first, _ := data[0].(map[string]interface{})
	return first, nil
}

// ListDir 列出目录
func (a *QuarkAPI) ListDir(pdirFid string, page int, size int) (map[string]interface{}, error) {
	u := fmt.Sprintf("%s/file/sort?pr=ucpro&fr=pc&uc_param_str=&pdir_fid=%s&_page=%d&_size=%d&_sort=file_type:asc,updated_at:desc",
		baseURL+apiVersion, pdirFid, page, size)
	return a.doJSONRequest("GET", u, nil)
}

// CreateDir 创建目录
func (a *QuarkAPI) CreateDir(pdirFid, dirName string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"pdir_fid":      pdirFid,
		"file_name":     dirName,
		"file_type":     0,
		"dir_init_lock": false,
	}
	u := fmt.Sprintf("%s/file?pr=ucpro&fr=pc&uc_param_str=", baseURL+apiVersion)
	return a.doJSONRequest("POST", u, payload)
}

// DeleteFile 删除文件
func (a *QuarkAPI) DeleteFile(fids []string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"action_type":  1,
		"filelist":     fids,
		"exclude_fids": []string{},
	}
	u := fmt.Sprintf("%s/file/delete?pr=ucpro&fr=pc&uc_param_str=", baseURL+apiVersion)
	return a.doJSONRequest("POST", u, payload)
}

// MoveFile 移动文件
func (a *QuarkAPI) MoveFile(fids []string, toPdirFid string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"action_type":  1,
		"exclude_fids": []string{},
		"filelist":     fids,
		"to_pdir_fid":  toPdirFid,
	}
	u := fmt.Sprintf("%s/file/move?pr=ucpro&fr=pc&uc_param_str=", baseURL+apiVersion)
	return a.doJSONRequest("POST", u, payload)
}

// RenameFile 重命名文件
func (a *QuarkAPI) RenameFile(fid, newName string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"fid":       fid,
		"file_name": newName,
	}
	u := fmt.Sprintf("%s/file/rename?pr=ucpro&fr=pc&uc_param_str=", baseURL+apiVersion)
	return a.doJSONRequest("POST", u, payload)
}

// User 获取用户信息
func (a *QuarkAPI) User() (map[string]interface{}, error) {
	u := fmt.Sprintf("%s/user?pr=ucpro&fr=pc&uc_param_str=", baseURL+apiVersion)
	return a.doJSONRequest("GET", u, nil)
}

// PartInfo XML 结构
type PartInfo struct {
	XMLName    xml.Name `xml:"Part"`
	PartNumber int      `xml:"PartNumber"`
	ETag       string   `xml:"ETag"`
}

type CompleteMultipartUpload struct {
	XMLName xml.Name `xml:"CompleteMultipartUpload"`
	Parts   []PartInfo
}
