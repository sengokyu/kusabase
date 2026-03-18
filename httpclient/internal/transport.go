package core

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Transport handles HTTP communication with the API.
type Transport struct {
	httpClient *http.Client
	baseURL    string
	saveCookie func(ctx context.Context, name, value string)
}

// NewTransport creates a new Transport.
// saveFn is called for each cookie received in a response.
func NewTransport(hc *http.Client, baseURL string, saveFn func(ctx context.Context, name, value string)) *Transport {
	return &Transport{httpClient: hc, baseURL: baseURL, saveCookie: saveFn}
}

func (t *Transport) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, t.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6.1 Safari/605.1.15")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (t *Transport) execute(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	for _, cookie := range resp.Cookies() {
		t.saveCookie(ctx, cookie.Name, cookie.Value)
	}
	return resp, nil
}

// DoJSON sends a request and decodes the JSON response into out.
// out may be nil if the response body is not needed.
func (t *Transport) DoJSON(ctx context.Context, method, path string, body, out any) error {
	req, err := t.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := t.execute(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// DoText sends a request and returns the response body as plain text.
func (t *Transport) DoText(ctx context.Context, method, path string, body any) (string, error) {
	req, err := t.newRequest(ctx, method, path, body)
	if err != nil {
		return "", err
	}

	resp, err := t.execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", parseError(resp)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
