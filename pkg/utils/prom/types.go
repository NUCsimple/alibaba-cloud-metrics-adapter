package prom

import (
	"encoding/json"
	"fmt"
)

// ErrorType is the type of the API error.
type ErrorType string

const (
	ErrBadData     ErrorType = "bad_data"
	ErrTimeout               = "timeout"
	ErrCanceled              = "canceled"
	ErrExec                  = "execution"
	ErrBadResponse           = "bad_response"
)

// Error is an error returned by the API.
type Error struct {
	Type ErrorType
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

// ResponseStatus is the type of response from the API: succeeded or error.
type ResponseStatus string

const (
	ResponseSucceeded ResponseStatus = "succeeded"
	ResponseError                    = "error"
)

// APIResponse represents the raw response returned by the API.
type APIResponse struct {
	// Status indicates whether this request was successful or whether it errored out.
	Status ResponseStatus `json:"status"`
	// Data contains the raw data response for this request.
	Data json.RawMessage `json:"data"`

	// ErrorType is the type of error, if this is an error response.
	ErrorType ErrorType `json:"errorType"`
	// Error is the error message, if this is an error response.
	Error string `json:"error"`
}