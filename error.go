package kusaclient

import "fmt"

// APIError represents a non-2xx HTTP response from the API.
type APIError struct {
	StatusCode int
	Namespace  string
	Key        string
}

func (e *APIError) Error() string {
	if e.Namespace != "" || e.Key != "" {
		return fmt.Sprintf("api error %d: %s:%s", e.StatusCode, e.Namespace, e.Key)
	}
	return fmt.Sprintf("api error %d", e.StatusCode)
}

// internal types for decoding error responses
type errorObject struct {
	NS  string `json:"ns"`
	Key string `json:"key"`
}

type errorResponse struct {
	Error errorObject `json:"error"`
}
