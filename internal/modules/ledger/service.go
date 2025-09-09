package ledger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Guizzs26/fintrack/pkg/clock"
	"github.com/google/uuid"
)

// AddTransactionParams holds all the required data for the AddTransactionToAccount use case
type AddTransactionParams struct {
	AccountID   uuid.UUID
	UserID      uuid.UUID
	CategoryID  *uuid.UUID
	Type        TransactionType
	Description string
	Observation string
	Amount      int64
	DueDate     time.Time
	PaidAt      *time.Time
}

// Service encapsulates the application's business logic (use cases) for the ledger module
type Service struct {
	accountRepo AccountRepository
	clock       clock.Clock
}

// NewService creates a new instance of the ledger Service
func NewLedgerService(accRepo AccountRepository, clock clock.Clock) *Service {
	return &Service{
		accountRepo: accRepo,
		clock:       clock,
	}
}

// CreateAccount is the use case for creating a new account
func (s *Service) CreateAccount(ctx context.Context, userID uuid.UUID, name string, includeInBalance bool) (*Account, error) {
	account, err := NewAccount(userID, name, includeInBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to create new account: %w", err)
	}

	if err := s.accountRepo.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save new account: %w", err)
	}

	return account, nil
}

// AddTransactionToAccount is the use case for adding a new transaction to an existing account
func (s *Service) AddTransactionToAccount(ctx context.Context, params AddTransactionParams) error {
	account, err := s.accountRepo.FindByID(ctx, params.AccountID)
	if err != nil {
		return fmt.Errorf("failed to find account to add transaction: %w", err)
	}

	// Enforce authorization rule: user can only modify their own account (FUTURE AUTHn/AUTHz)
	if account.UserID != params.UserID {
		return errors.New("user does not have permission to access this account")
	}

	err = account.AddTransaction(
		params.Type,
		params.Description,
		params.Observation,
		params.Amount,
		params.CategoryID,
		params.DueDate,
		params.PaidAt,
		s.clock,
	)
	if err != nil {
		return fmt.Errorf("failed to add transaction: %w", err)
	}

	if err := s.accountRepo.Save(ctx, account); err != nil {
		return fmt.Errorf("failed to save account after adding transaction: %w", err)
	}

	return nil
}
