package core

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error represents a non-2xx HTTP response from the API.
type Error struct {
	StatusCode int
	Namespace  string
	Key        string
}

func (e *Error) Error() string {
	if e.Namespace != "" || e.Key != "" {
		return fmt.Sprintf("api error %d: %s:%s", e.StatusCode, e.Namespace, e.Key)
	}
	return fmt.Sprintf("api error %d", e.StatusCode)
}

type errorObject struct {
	NS  string `json:"ns"`
	Key string `json:"key"`
}

type errorResponse struct {
	Error errorObject `json:"error"`
}

func parseError(resp *http.Response) error {
	apiErr := &Error{StatusCode: resp.StatusCode}
	var errResp errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		apiErr.Namespace = errResp.Error.NS
		apiErr.Key = errResp.Error.Key
	}
	return apiErr
}
