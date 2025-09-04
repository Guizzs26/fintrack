package ledger

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// LedgerHandler holds dependencies for ledger-related HTTP handlers
type LedgerHandler struct {
	ledgerService *Service
}

// NewLedgerHandler creates a new instance of LedgerHandler
func NewLedgerHandler(ledgerService *Service) *LedgerHandler {
	return &LedgerHandler{ledgerService: ledgerService}
}

// RegisterRoutes sets up the API routes for the ledger module
func (h *LedgerHandler) RegisterRoutes(e *echo.Echo) {
	apiGroup := e.Group("/api/v1")

	apiGroup.POST("/accounts", h.CreateAccountHandler)
}

// CreateAccountRequest defines the expected JSON body for creating a new account
type CreateAccountRequest struct {
	Name                    string `json:"name" validate:"required,min=1,max=100"`
	IncludeInOverallBalance *bool  `json:"include_in_overall_balance,omitempty"`
}

// AccountResponse defines the structure of an account returned by the API
type AccountResponse struct {
	ID                      uuid.UUID `json:"id"`
	UserID                  uuid.UUID `json:"user_id"`
	Name                    string    `json:"name"`
	IncludeInOverallBalance bool      `json:"include_in_overall_balance"`
}

// CreateAccountHandler handles the HTTP request for creating a new account
func (h *LedgerHandler) CreateAccountHandler(c echo.Context) error {
	var req CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}

	includeInBalance := true
	if req.IncludeInOverallBalance != nil {
		includeInBalance = *req.IncludeInOverallBalance
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	account, err := h.ledgerService.CreateAccount(c.Request().Context(), mockUserID, req.Name, includeInBalance)
	if err != nil {
		if errors.Is(err, ErrAccountNameRequired) {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
	}

	response := AccountResponse{
		ID:                      account.ID,
		UserID:                  account.UserID,
		Name:                    account.Name,
		IncludeInOverallBalance: account.IncludeInOverallBalance,
	}

	return c.JSON(http.StatusCreated, response)
}
