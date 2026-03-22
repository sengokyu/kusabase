package core

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// SaveCookiesFunc is called with all cookies received in a response.
type SaveCookiesFunc func(ctx context.Context, cookies []*http.Cookie)

// Transport handles HTTP communication with the API.
type Transport struct {
	httpClient  *http.Client
	baseURL     string
	saveCookies SaveCookiesFunc
}

// NewTransport creates a new Transport.
// saveFn is called with all cookies received in a response.
func NewTransport(hc *http.Client, baseURL string, saveFn SaveCookiesFunc) *Transport {
	return &Transport{httpClient: hc, baseURL: baseURL, saveCookies: saveFn}
}

func (t *Transport) buildRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, t.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6.1 Safari/605.1.15")
	req.Header.Set("Referer", t.baseURL)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	return req, nil
}

func (t *Transport) newGetRequest(ctx context.Context, path string) (*http.Request, error) {
	return t.buildRequest(ctx, "GET", path, nil)
}

func (t *Transport) newPostRequest(ctx context.Context, path string, body any) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := t.buildRequest(ctx, "POST", path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (t *Transport) do(req *http.Request, handle func(*http.Response) error) error {
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	if cookies := resp.Cookies(); len(cookies) > 0 {
		t.saveCookies(req.Context(), cookies)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp)
	}
	return handle(resp)
}

func (t *Transport) doJSON(req *http.Request, out any) error {
	req.Header.Set("Accept", "application/json")
	return t.do(req, func(resp *http.Response) error {
		if out != nil {
			return json.NewDecoder(resp.Body).Decode(out)
		}
		return nil
	})
}

func (t *Transport) doText(req *http.Request) (string, error) {
	var result string
	err := t.do(req, func(resp *http.Response) error {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		result = string(b)
		return nil
	})
	return result, err
}

// GetJSON sends a GET request and decodes the JSON response into out.
// out may be nil if the response body is not needed.
func (t *Transport) GetJSON(ctx context.Context, path string, out any) error {
	req, err := t.newGetRequest(ctx, path)
	if err != nil {
		return err
	}
	return t.doJSON(req, out)
}

// PostJSON sends a POST request and decodes the JSON response into out.
// out may be nil if the response body is not needed.
func (t *Transport) PostJSON(ctx context.Context, path string, body, out any) error {
	req, err := t.newPostRequest(ctx, path, body)
	if err != nil {
		return err
	}
	return t.doJSON(req, out)
}

// GetText sends a GET request and returns the response body as plain text.
func (t *Transport) GetText(ctx context.Context, path string) (string, error) {
	req, err := t.newGetRequest(ctx, path)
	if err != nil {
		return "", err
	}
	return t.doText(req)
}

// PostText sends a POST request and returns the response body as plain text.
func (t *Transport) PostText(ctx context.Context, path string, body any) (string, error) {
	req, err := t.newPostRequest(ctx, path, body)
	if err != nil {
		return "", err
	}
	return t.doText(req)
}
