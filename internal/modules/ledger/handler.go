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
	clock         clock.Clock
}

// NewLedgerHandler creates a new instance of LedgerHandler
func NewLedgerHandler(ledgerService *Service, clock clock.Clock) *LedgerHandler {
	return &LedgerHandler{
		ledgerService: ledgerService,
		clock:         clock,
	}
}

// RegisterRoutes sets up the API routes for the ledger module
func (h *LedgerHandler) RegisterRoutes(apiRouteGroup *echo.Group) {
	accountsGroup := apiRouteGroup.Group("/accounts")

	accountsGroup.POST("", h.createAccountHandler)
	accountsGroup.POST("/:id/transactions", h.addTransactionHandler)
	accountsGroup.PUT("/:id", h.updateAccountHandler)
	accountsGroup.POST("/:id/balance-adjustment", h.accountBalanceAdjustmentHandler)
	accountsGroup.DELETE("/:id", h.archiveAccountHandler)
	accountsGroup.POST("/:id/unarchive", h.unarchiveAccountHandler)
	accountsGroup.GET("/:id", h.findAccountByIDHandler)
	accountsGroup.GET("", h.findAccountsByUserIDHandler)
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

// UpdateAccountRequest defines the expected JSON body for updating an account
type UpdateAccountRequest struct {
	Name                    *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	IncludeInOverallBalance *bool   `json:"include_in_overall_balance,omitempty"`
}

// BalanceAdjustmentRequest defines the expected JSON body for adjust the account balance
type BalanceAdjustmentRequest struct {
	NewBalance int64 `json:"new_balance" validate:"required"`
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

// AccountSummaryResponse defines a summary view of an account for list endpoints
type AccountSummaryResponse struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	RealBalance      int64     `json:"real_balance"`
	ProjectedBalance int64     `json:"projected_balance"`
}

// CurrentMonthFlowSummary details the income, expenses and net (balance) result of the current month
type CurrentMonthFlowSummary struct {
	Income  int64 `json:"income"`
	Expense int64 `json:"expense"`
	NetFlow int64 `json:"net_flow"` // income - expense
}

// AccountListResponse is the DTO for the response listing all the user's accounts
// Includes the list of accounts and the calculated overall balances
type AccountListResponse struct {
	OverallRealBalance      int64                    `json:"overall_real_balance"`
	OverallProjectedBalance int64                    `json:"overall_projected_balance"`
	CurrentMonthFlow        CurrentMonthFlowSummary  `json:"current_month_flow"`
	Accounts                []AccountSummaryResponse `json:"accounts"`
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

// updateAccountHandler handles HTTP request for update a existing account
func (h *LedgerHandler) updateAccountHandler(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id format")
	}

	var req UpdateAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body format")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	if req.Name == nil && req.IncludeInOverallBalance == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "at least one field must be provided for update")
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	params := UpdateAccountParams{
		AccountID:               accountID,
		UserID:                  mockUserID,
		Name:                    req.Name,
		IncludeInOverallBalance: req.IncludeInOverallBalance,
	}

	account, err := h.ledgerService.UpdateAccount(c.Request().Context(), params)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusOK, toAccountResponse(account))
}

// archiveAccountHandler handles HTTP request for archive a existing account
func (h *LedgerHandler) archiveAccountHandler(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account ID format")
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	if err := h.ledgerService.ArchiveAccount(c.Request().Context(), mockUserID, accountID); err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusNoContent, nil)
}

// unarchiveAccountHandler handles HTTP request for unarchived a archived account
func (h *LedgerHandler) unarchiveAccountHandler(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid account ID format")
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	account, err := h.ledgerService.UnarchiveAccount(c.Request().Context(), mockUserID, accountID)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusOK, toAccountResponse(account))
}

// accountBalanceAdjustmentHandler handles HTTP request for adjust the balance of an existing account
func (h *LedgerHandler) accountBalanceAdjustmentHandler(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account ID format")
	}

	var req BalanceAdjustmentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body format")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	params := BalanceAdjustmentParams{
		AccountID:  accountID,
		UserID:     mockUserID,
		NewBalance: req.NewBalance,
	}

	account, err := h.ledgerService.AdjustAccountBalance(c.Request().Context(), params)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusOK, toAccountDetailResponse(account, h.clock))
}

// findAccountByID handles the HTTP request for finding a account by id
func (h *LedgerHandler) findAccountByIDHandler(c echo.Context) error {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id format")
	}

	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	account, err := h.ledgerService.FindAccountByID(c.Request().Context(), mockUserID, accountID)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusOK, toAccountDetailResponse(account, h.clock))
}

// findAccountsByUserIDHandler handles the HTTP request for finding the account(s) by the user id
func (h *LedgerHandler) findAccountsByUserIDHandler(c echo.Context) error {
	mockUserID, _ := uuid.Parse("7e57d19c-5953-433c-9b57-d3d8e1f3b8b8")
	accounts, err := h.ledgerService.FindAccountsByUserID(c.Request().Context(), mockUserID)
	if err != nil {
		return err
	}

	return httpx.SendSuccess(c, http.StatusOK, toAccountListResponse(accounts, h.clock))
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
func toAccountDetailResponse(a *Account, clock clock.Clock) AccountDetailResponse {
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

	return AccountDetailResponse{
		ID:                      a.ID,
		Name:                    a.Name,
		RealBalance:             a.RealBalance(clock),
		ProjectedBalance:        a.ProjectedBalance(),
		IncludeInOverallBalance: a.IncludeInOverallBalance,
		Transactions:            txResponses,
	}
}

// toAccountListResponse maps a slice of Accounts from the domain to the public DTO AccountListResponse
// It is responsible for calculating the overall balances and cash flow for the *current month*
func toAccountListResponse(accounts []*Account, clock clock.Clock) AccountListResponse {
	now := clock.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)

	var overallRealBalance int64 = 0
	var overallProjectedBalance int64 = 0
	var currentMonthIncome int64 = 0
	var currentMonthExpense int64 = 0
	accountSummaries := make([]AccountSummaryResponse, len(accounts))

	for i, acc := range accounts {
		realBalance := acc.RealBalance(clock)
		projectedBalance := acc.ProjectedBalance()

		// A .presentation business logic: overall balance calculation (regardless of the period)
		if acc.IncludeInOverallBalance {
			overallRealBalance += realBalance
			overallProjectedBalance += projectedBalance
		}

		// calculate the current month's flow
		if acc.IncludeInOverallBalance {
			for _, tx := range acc.Transactions() {
				// The transaction only enters the monthly flow if:
				// 1. It was paid/completed (PaidAt is not null)
				// 2. The payment date is within the current month's range
				if tx.PaidAt != nil && !tx.PaidAt.Before(startOfMonth) && tx.PaidAt.Before(startOfNextMonth) {
					switch tx.Type {
					case Income, Adjustment:
						currentMonthIncome += tx.Amount
					case Expense:
						currentMonthExpense += tx.Amount
					}
				}
			}
		}

		accountSummaries[i] = AccountSummaryResponse{
			ID:               acc.ID,
			Name:             acc.Name,
			RealBalance:      realBalance,
			ProjectedBalance: projectedBalance,
		}
	}

	return AccountListResponse{
		OverallRealBalance:      overallRealBalance,
		OverallProjectedBalance: overallProjectedBalance,
		CurrentMonthFlow: CurrentMonthFlowSummary{
			Income:  currentMonthIncome,
			Expense: currentMonthExpense,
			NetFlow: currentMonthIncome + currentMonthExpense},
		Accounts: accountSummaries,
	}
}
