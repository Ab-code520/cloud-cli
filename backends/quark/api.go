package quark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Ab-code520/cloud-cli/utils"
)

const (
	baseURL      = "https://drive-pc.quark.cn"
	apiPrecreate = "/1/clouddrive/file/upload/pre"
	apiAuth      = "/1/clouddrive/file/upload/auth"
	apiFinish    = "/1/clouddrive/file/upload/finish"
	apiList      = "/1/clouddrive/file/sort"
	apiDownload  = "/1/clouddrive/file/download"
)

type QuarkAPI struct {
	client  *http.Client
	cookie  string
	cookies map[string]string
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
		client:  &http.Client{Timeout: 60 * time.Second},
		cookie:  cookie,
		cookies: cookies,
	}
}

func (a *QuarkAPI) request(ctx context.Context, method, path string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", a.cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := a.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var baseResp struct {
		Status  int             `json:"status"`
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respData, &baseResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if baseResp.Status != 200 || baseResp.Code != 0 {
		return fmt.Errorf("api error %d: %s", baseResp.Code, baseResp.Message)
	}

	if result != nil {
		return json.Unmarshal(baseResp.Data, result)
	}
	return nil
}

type PrecreateReq struct {
	PdirFid   string `json:"pdir_fid"`
	FileName  string `json:"file_name"`
	Size      int64  `json:"size"`
	Sha1      string `json:"sha1"`
	ChunkSize int64  `json:"chunk_size"`
}

type PrecreateResp struct {
	UploadID    string `json:"upload_id"`
	RapidUpload bool   `json:"rapid_upload"`
	PartSize    int64  `json:"part_size"`
	FileID      string `json:"fid"`
}

func (a *QuarkAPI) Precreate(ctx context.Context, req *PrecreateReq) (*PrecreateResp, error) {
	var resp PrecreateResp
	err := a.request(ctx, http.MethodPost, apiPrecreate, req, &resp)
	return &resp, err
}

type AuthReq struct {
	UploadID string `json:"upload_id"`
}

type AuthResp struct {
	Endpoint string `json:"endpoint"`
}

func (a *QuarkAPI) GetUploadAuth(ctx context.Context, uploadID string) (*AuthResp, error) {
	req := AuthReq{UploadID: uploadID}
	var resp AuthResp
	err := a.request(ctx, http.MethodPost, apiAuth, req, &resp)
	return &resp, err
}

func (a *QuarkAPI) UploadPart(ctx context.Context, ossURL string, data []byte, partNumber int, uploadID string) (string, error) {
	if safe, reason := utils.IsURLSafe(ossURL); !safe {
		return "", fmt.Errorf("ssrf blocked: %s", reason)
	}

	finalURL := fmt.Sprintf("%s?partNumber=%d&uploadId=%s", ossURL, partNumber, uploadID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, finalURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := a.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("upload part failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload part failed with status %d", resp.StatusCode)
	}

	// Safe ETag extraction
	etag := resp.Header.Get("ETag")
	etag = strings.ReplaceAll(etag, string(rune(34)), "")
	return etag, nil
}

type FinishReq struct {
	UploadID string `json:"upload_id"`
	Parts    []struct {
		PartNumber int    `json:"part_number"`
		ETag       string `json:"etag"`
	} `json:"parts"`
}

func (a *QuarkAPI) FinishUpload(ctx context.Context, req *FinishReq) error {
	return a.request(ctx, http.MethodPost, apiFinish, req, nil)
}

type ListReq struct {
	PdirFid string `json:"pdir_fid"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
	Sort    string `json:"sort"`
	Order   string `json:"order"`
}

type FileItem struct {
	Fid       string `json:"fid"`
	FileName  string `json:"file_name"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"is_dir"`
	UpdatedAt int64  `json:"updated_at"`
}

type ListResp struct {
	List  []FileItem `json:"list"`
	Total int        `json:"total"`
}

func (a *QuarkAPI) List(ctx context.Context, req *ListReq) (*ListResp, error) {
	var resp ListResp
	err := a.request(ctx, http.MethodPost, apiList, req, &resp)
	return &resp, err
}

type DownloadReq struct {
	Fids []string `json:"fids"`
}

type DownloadItem struct {
	DownloadURL string `json:"download_url"`
	Fid         string `json:"fid"`
}

type DownloadResp struct {
	List []DownloadItem `json:"list"`
}

func (a *QuarkAPI) GetDownloadURL(ctx context.Context, fid string) (string, error) {
	req := DownloadReq{Fids: []string{fid}}
	var resp DownloadResp
	err := a.request(ctx, http.MethodPost, apiDownload, req, &resp)
	if err != nil {
		return "", err
	}
	if len(resp.List) == 0 {
		return "", fmt.Errorf("no download url found")
	}
	return resp.List[0].DownloadURL, nil
}

// ═══════════════════════════════════════════
// QR Login API
// ═══════════════════════════════════════════

const (
	qrBaseURL    = "https://uop.quark.cn"
	qrGenerate   = "/api/v1/auth/qrcode/generate"
	qrQuery      = "/api/v1/auth/qrcode/query"
	qrAuthorize  = "https://su.quark.cn/4_e3f80838a74530001d816a5628a6addd"
)

// QRTokenResp is the response from generating a QR code.
type QRTokenResp struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// GenerateQR creates a new QR login session and returns the token + login URL.
func (a *QuarkAPI) GenerateQR(ctx context.Context) (*QRTokenResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, qrBaseURL+qrGenerate, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("generate QR failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var baseResp struct {
		Status  int             `json:"status"`
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respData, &baseResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if baseResp.Status != 200 || baseResp.Code != 0 {
		return nil, fmt.Errorf("api error %d: %s", baseResp.Code, baseResp.Message)
	}

	var qrResp QRTokenResp
	if err := json.Unmarshal(baseResp.Data, &qrResp); err != nil {
		return nil, fmt.Errorf("parse QR data: %w", err)
	}
	return &qrResp, nil
}

// QRStatus is the polling result for QR login status.
type QRStatus string

const (
	QRNotScanned   QRStatus = "NOT_SCANNED"
	QRScanned      QRStatus = "SCANNED"
	QRConfirmed    QRStatus = "CONFIRMED"
	QRExpired      QRStatus = "EXPIRED"
	QRUnknownState QRStatus = "UNKNOWN"
)

// QRQueryResp is the response from polling QR status.
type QRQueryResp struct {
	Status      QRStatus `json:"status"`
	Cookie      string   `json:"cookie"`
	RedirectURL string   `json:"redirect_url"`
	Message     string   `json:"message"`
}

// QueryQRStatus polls the QR login status.
// Returns the status and optional cookie/redirect_url when confirmed.
func (a *QuarkAPI) QueryQRStatus(ctx context.Context, token string) (*QRQueryResp, error) {
	url := fmt.Sprintf("%s%s?token=%s", qrBaseURL, qrQuery, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("query QR status failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var baseResp struct {
		Status  int             `json:"status"`
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respData, &baseResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	var qrResult QRQueryResp

	// Quark uses status 200001 for not scanned, 200003 for scanned, 200004 for confirmed/expired
	switch baseResp.Code {
	case 0:
		// Parsed as confirmed, try to extract cookie from data
		if baseResp.Data != nil {
			json.Unmarshal(baseResp.Data, &qrResult)
		}
		qrResult.Status = QRConfirmed
		return &qrResult, nil
	case 200001:
		qrResult.Status = QRNotScanned
		return &qrResult, nil
	case 200003:
		qrResult.Status = QRScanned
		return &qrResult, nil
	case 200004:
		qrResult.Status = QRExpired
		return &qrResult, nil
	default:
		qrResult.Status = QRUnknownState
		qrResult.Message = baseResp.Message
		return &qrResult, nil
	}
}
