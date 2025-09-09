package ledger

import (
	"net/http"
	"time"

	"github.com/Guizzs26/fintrack/pkg/clock"
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
	accountsGroup.POST("/:id/transactions", h.addTransactionHandler)
	accountsGroup.GET("/:id", h.findAccountByID)
}

// CreateAccountRequest defines the expected JSON body for creating a new account
type CreateAccountRequest struct {
	Name                    string `json:"name" validate:"required,min=1,max=100"`
	IncludeInOverallBalance *bool  `json:"include_in_overall_balance,omitempty"`
}

// AddTransactionRequest defines the expected JSON body for creating a transaction for an account
type AddTransactionRequest struct {
	Type        TransactionType `json:"type" validate:"required"`
	Description string          `json:"description" validate:"required,min=1,max=100"`
	Observation string          `json:"observation,omitempty" validate:"max=2500"`
	Amount      int64           `json:"amount" validate:"required,ne=0"`
	DueDate     time.Time       `json:"due_date" validate:"required"`
	PaidAt      *time.Time      `json:"paid_at,omitempty"`
	CategoryID  *uuid.UUID      `json:"category_id,omitempty"`
}

// TransactionResponse defines the structure of an transaction returned by the API
type TransactionResponse struct {
	ID          uuid.UUID       `json:"id"`
	Type        TransactionType `json:"type"`
	Description string          `json:"description"`
	Amount      int64           `json:"amount"`
	DueDate     time.Time       `json:"due_date"`
	PaidAt      *time.Time      `json:"paid_at,omitempty"`
}

// AccountResponse defines the structure of an account returned by the API
type AccountResponse struct {
	ID                      uuid.UUID `json:"id"`
	UserID                  uuid.UUID `json:"user_id"`
	Name                    string    `json:"name"`
	IncludeInOverallBalance bool      `json:"include_in_overall_balance"`
}

// AccountDetailResponse defines the structure of an detailed account + transaction response returned by the API
type AccountDetailResponse struct {
	ID                      uuid.UUID             `json:"id"`
	Name                    string                `json:"name"`
	RealBalance             int64                 `json:"real_balance"`
	ProjectedBalance        int64                 `json:"projected_balance"`
	IncludeInOverallBalance bool                  `json:"include_in_overall_balance"`
	Transactions            []TransactionResponse `json:"transactions"`
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

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	account, err := h.ledgerService.CreateAccount(c.Request().Context(), mockUserID, req.Name, includeInBalance)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusCreated, toAccountResponse(account))
}

// addTransactionHandler handles the HTTP request for creating a new transaction
func (h *LedgerHandler) addTransactionHandler(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id format")
	}

	var req AddTransactionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body format")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	params := AddTransactionParams{
		AccountID:   accountID,
		UserID:      mockUserID, // In the future, this will come from JWT/middleware
		Type:        req.Type,
		Description: req.Description,
		Observation: req.Observation,
		Amount:      req.Amount,
		DueDate:     req.DueDate,
		PaidAt:      req.PaidAt,
		CategoryID:  req.CategoryID,
	}

	if err := h.ledgerService.AddTransactionToAccount(c.Request().Context(), params); err != nil {
		return err
	}

	// For a POST that creates a sub-resource, 204 No Content is a valid and efficient response
	return httpx.SendSuccess(c, http.StatusNoContent, nil)
}

// findAccountByID handles the HTTP request for finding a account by id
func (h *LedgerHandler) findAccountByID(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id format")
	}

	account, err := h.ledgerService.FindAccountByID(c.Request().Context(), accountID)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusOK, toAccountDetailResponse(account))
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

// toAccountDetailResponse maps the internal Account domain model to the public AccountDetailResponse DTO
func toAccountDetailResponse(a *Account) AccountDetailResponse {
	txs := a.Transactions()
	txResponses := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = TransactionResponse{
			ID:          tx.ID,
			Type:        tx.Type,
			Description: tx.Description,
			Amount:      tx.Amount,
			DueDate:     tx.DueDate,
			PaidAt:      tx.PaidAt,
		}
	}

	clock := clock.SystemClock{}
	return AccountDetailResponse{
		ID:                      a.ID,
		Name:                    a.Name,
		RealBalance:             a.RealBalance(clock),
		ProjectedBalance:        a.ProjectedBalance(),
		IncludeInOverallBalance: a.IncludeInOverallBalance,
		Transactions:            txResponses,
	}
}
