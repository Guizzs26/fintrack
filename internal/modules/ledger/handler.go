package ledger

import (
	"net/http"

	"github.com/Guizzs26/fintrack/pkg/httpx"
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
func (h *LedgerHandler) RegisterRoutes(apiRouteGroup *echo.Group) {
	accountsGroup := apiRouteGroup.Group("/accounts")

	// POST /api/v1/accounts
	accountsGroup.POST("", h.createAccountHandler)
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

// createAccountHandler handles the HTTP request for creating a new account
func (h *LedgerHandler) createAccountHandler(c echo.Context) error {
	var req CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body format")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	includeInBalance := true
	if req.IncludeInOverallBalance != nil {
		includeInBalance = *req.IncludeInOverallBalance
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b")
	account, err := h.ledgerService.CreateAccount(c.Request().Context(), mockUserID, req.Name, includeInBalance)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusCreated, toAccountResponse(account))
}

// toAccountResponse maps the internal Account domain model to the public AccountResponse DTO
func toAccountResponse(a *Account) AccountResponse {
	return AccountResponse{
		ID:                      a.ID,
		UserID:                  a.UserID,
		Name:                    a.Name,
		IncludeInOverallBalance: a.IncludeInOverallBalance,
	}
}
