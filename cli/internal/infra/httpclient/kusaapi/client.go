package kusaapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// i18nText はロケール別テキストを保持する JSON パース用型。
type i18nText struct {
	JA string `json:"ja"`
	EN string `json:"en"`
}

// Resolve は日本語を優先してテキストを返す。
func (t i18nText) Resolve() string {
	if t.JA != "" {
		return t.JA
	}
	return t.EN
}

// Client is the HTTP client for the kusa API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	cookieJar  *PersistentCookieJar
	debugf     func(string, ...interface{})
}

// NewClient creates a new API Client. cacheDir is used for cookie persistence.
func NewClient(baseURL, cacheDir string, debugf func(string, ...interface{})) (*Client, error) {
	jar, err := NewPersistentCookieJar(baseURL, cacheDir)
	if err != nil {
		return nil, fmt.Errorf("cookie jar の初期化に失敗しました: %w", err)
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Jar: jar, Timeout: 60 * time.Second},
		cookieJar:  jar,
		debugf:     debugf,
	}, nil
}

// Close saves the cookie jar to disk.
func (c *Client) Close() error {
	return c.cookieJar.Save()
}

// checkStatus は 401/403 を ErrNotLoggedIn に変換し、非 200 はエラーを返す。
// path はエラーメッセージのコンテキスト（エンドポイントパス）に使用する。
func checkStatus(resp *http.Response, path string) error {
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ports.ErrNotLoggedIn
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP %d", path, resp.StatusCode)
	}
	return nil
}

// getJSON は GET リクエストを送信し、レスポンスを JSON デコードする。
func (c *Client) getJSON(ctx context.Context, path string, v any) error {
	resp, err := c.get(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, path); err != nil {
		return err
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}
	return nil
}

func (c *Client) post(ctx context.Context, path string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("リクエストのシリアライズに失敗しました: %w", err)
	}
	c.debugf("POST %s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエストの送信に失敗しました: %w", err)
	}
	if err := checkStatus(resp, path); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	c.debugf("GET %s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエストの送信に失敗しました: %w", err)
	}
	return res, nil
}

// Ensure Client satisfies the interface at compile time.
var _ ports.ExternalAPIClient = (*Client)(nil)
