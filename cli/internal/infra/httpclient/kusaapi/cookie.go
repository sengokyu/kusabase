package kusaapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// PersistentCookieJar is an http.CookieJar that persists cookies to disk.
type PersistentCookieJar struct {
	jar      *cookiejar.Jar // 標準ライブラリのインメモリ cookie jar
	baseURL  *url.URL       // Cookie を取得する際の基準 URL
	filePath string         // Cookie の保存先ファイルパス
}

// storedCookie は Cookie をファイルに保存するための JSON シリアライズ用構造体。
type storedCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Path     string    `json:"path"`
	Domain   string    `json:"domain"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"httpOnly"`
}

// NewPersistentCookieJar creates a PersistentCookieJar, loading any existing cookies from disk.
func NewPersistentCookieJar(baseURL, cacheDir string) (*PersistentCookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookie jar の作成に失敗しました: %w", err)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("URL のパースに失敗しました: %w", err)
	}

	pcj := &PersistentCookieJar{
		jar:      jar,
		baseURL:  u,
		filePath: filepath.Join(cacheDir, "cookies.bin"),
	}

	if err := pcj.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "警告: Cookie の読み込みに失敗しました。再ログインが必要な場合があります。\n")
	}

	return pcj, nil
}

func (j *PersistentCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.jar.SetCookies(u, cookies)
}

func (j *PersistentCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.jar.Cookies(u)
}

// Save persists the current cookies to disk.
func (j *PersistentCookieJar) Save() error {
	cookies := j.jar.Cookies(j.baseURL)

	stored := make([]storedCookie, 0, len(cookies))
	for _, c := range cookies {
		stored = append(stored, storedCookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		})
	}

	data, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("Cookie のシリアライズに失敗しました: %w", err)
	}

	dir := filepath.Dir(j.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	if err := os.WriteFile(j.filePath, data, 0600); err != nil {
		return fmt.Errorf("Cookie の保存に失敗しました: %w", err)
	}

	return nil
}

// Load restores cookies from disk.
func (j *PersistentCookieJar) Load() error {
	data, err := os.ReadFile(j.filePath)
	if err != nil {
		return err
	}

	var stored []storedCookie
	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("Cookie のデシリアライズに失敗しました: %w", err)
	}

	cookies := make([]*http.Cookie, 0, len(stored))
	for _, s := range stored {
		cookies = append(cookies, &http.Cookie{
			Name:     s.Name,
			Value:    s.Value,
			Path:     s.Path,
			Domain:   s.Domain,
			Expires:  s.Expires,
			Secure:   s.Secure,
			HttpOnly: s.HttpOnly,
		})
	}

	j.jar.SetCookies(j.baseURL, cookies)
	return nil
}
