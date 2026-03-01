package kusaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// newRequest builds an *http.Request with an optional JSON body.
func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// execute sends req, persists any updated session cookie, and returns the response.
func (c *Client) execute(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "next-session" {
			_ = c.store.Save(ctx, "next-session", cookie.Value)
		}
	}
	return resp, nil
}

// doJSON makes a request and decodes a JSON response body into out.
// out may be nil if the response body is not needed.
func (c *Client) doJSON(ctx context.Context, method, path string, body, out any) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.execute(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// doText makes a request and returns the response body as plain text.
func (c *Client) doText(ctx context.Context, method, path string, body any) (string, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return "", err
	}

	resp, err := c.execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", parseAPIError(resp)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseAPIError(resp *http.Response) error {
	apiErr := &APIError{StatusCode: resp.StatusCode}
	var errResp errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		apiErr.Namespace = errResp.Error.NS
		apiErr.Key = errResp.Error.Key
	}
	return apiErr
}
