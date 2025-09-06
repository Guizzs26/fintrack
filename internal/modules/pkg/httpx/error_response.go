package httpx

import "github.com/labstack/echo/v4"

// APIError is the standard wrapper for all error API responses (4xx and 5xx status codes)
// It provides a consistent, machine-readable format for clients to handle failures
type APIError struct {
	Code    string `json:"code"`              // A machine-readable error code (e.g., "VALIDATION_ERROR", "RESOURCE_NOT_FOUND")
	Message string `json:"message"`           // A human-readable message intended for the developer consuming the API
	Details any    `json:"details,omitempty"` // An optional field for providing more specific context, like a slice of validation errors
}

// NewAPIError creates a new APIError response structure
func NewAPIError(code, message string, details any) APIError {
	return APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// SendAPIError is a helper function to standardize sending error JSON responses
func SendAPIError(c echo.Context, httpStatus int, err APIError) error {
	return c.JSON(httpStatus, err)
}
