// 文档解析服务（docling，Python/FastAPI :8300）的 HTTP 客户端。
// Go 网关作为唯一入口,对 docling 做薄代理:注入调用者省份(org)、统一鉴权/审计,
// 响应原样透传(不在 Go 侧重建 docling 的数据形状)。
package docling

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		// OCR 逐页调外部视觉 API,可能较慢,给足超时。
		http: &http.Client{Timeout: 300 * time.Second},
	}
}

// Parse 转发上传文件到 docling /api/v1/parse(multipart),带上 org 与分类。
// 返回 docling 的原始响应体与状态码。
func (c *Client) Parse(ctx context.Context, filename string, content []byte,
	docCategory, orgID string, saveToDB bool) ([]byte, int, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, 0, err
	}
	if _, err := fw.Write(content); err != nil {
		return nil, 0, err
	}
	_ = w.WriteField("doc_category", docCategory)
	_ = w.WriteField("save_to_db", strconv.FormatBool(saveToDB))
	if orgID != "" {
		_ = w.WriteField("org_id", orgID)
	}
	if err := w.Close(); err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/parse", &body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.do(req)
}

// ParseResult docling /parse 的结构化响应。
type ParseResult struct {
	Filename    string              `json:"filename"`
	DocType     string              `json:"doc_type"`
	TextContent string              `json:"text_content"`
	Tables      []map[string]any    `json:"tables"`
	Entities    map[string][]string `json:"entities"`
	Summary     string              `json:"summary"`
	PageCount   int                 `json:"page_count"`
}

// ParseDoc 解析文件并解码为结构化结果（save_to_db=false：持久化由 Go 网关统一负责）。
func (c *Client) ParseDoc(ctx context.Context, filename string, content []byte, docCategory string) (*ParseResult, error) {
	data, status, err := c.Parse(ctx, filename, content, docCategory, "", false)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("docling 解析失败(HTTP %d): %s", status, truncate(string(data), 300))
	}
	var r ParseResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("docling 响应解析失败: %w", err)
	}
	return &r, nil
}

// ExtractField GLM 结构化提取的单个字段。
type ExtractField struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Value      string  `json:"value"`
	Unit       string  `json:"unit"`
	Confidence float64 `json:"confidence"`
}

// Extract 调 docling /api/v1/extract：按文档类型从全文定向提取结构化字段。
func (c *Client) Extract(ctx context.Context, text, docType string) ([]ExtractField, error) {
	payload, _ := json.Marshal(map[string]string{"text": text, "doc_type": docType})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/extract", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	data, status, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("docling 提取失败(HTTP %d): %s", status, truncate(string(data), 300))
	}
	var out struct {
		Fields []ExtractField `json:"fields"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("docling 提取响应解析失败: %w", err)
	}
	return out.Fields, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (c *Client) do(req *http.Request) ([]byte, int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("docling 服务不可达: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}
