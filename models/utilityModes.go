package models

// SuccessResponse represents a success response body
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response body
type ErrorResponse struct {
	Message string `json:"message"`
}
