package httpx

import "github.com/labstack/echo/v4"

// Success is the standard wrapper for all successful API responses (2xx status codes)
// The Data field is generic and can hold any type of payload, such as a single resource object or a slice of resources
type Success struct {
	Data any `json:"data"`
}

// NewSuccess creates a new Success response structure, wrapping the provided data
func NewSuccess(data any) *Success {
	return &Success{Data: data}
}

// SendSuccess is a helper function to standardize sending successful JSON responses
// It wraps the data in the Success structure and sends it with the given status code
func SendSuccess(c echo.Context, code int, data interface{}) error {
	return c.JSON(code, NewSuccess(data))
}
