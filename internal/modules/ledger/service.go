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

// UpdateAccountParams hols all the required data for UpdateAccount use case
type UpdateAccountParams struct {
	AccountID               uuid.UUID
	UserID                  uuid.UUID
	Name                    *string
	IncludeInOverallBalance *bool
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

// UpdateAccount is the use case for update an existing account
func (s *Service) UpdateAccount(ctx context.Context, params UpdateAccountParams) (*Account, error) {
	account, err := s.FindAccountByID(ctx, params.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account to update the account: %w", err)
	}

	if account.UserID != params.UserID {
		return nil, errors.New("user does not have permission to access this account")
	}

	if params.Name != nil {
		if err := account.ChangeName(*params.Name); err != nil {
			return nil, fmt.Errorf("failed to update account name: %w", err)
		}
	}

	if params.IncludeInOverallBalance != nil {
		if *params.IncludeInOverallBalance {
			if err := account.EnableOverallBalance(); err != nil {
				if !errors.Is(err, ErrAccountAlreadyIncluded) {
					return nil, fmt.Errorf("failed to include account in overall balance: %w", err)
				}
			}
		} else {
			if err := account.DisableOverallBalance(); err != nil {
				if !errors.Is(err, ErrAccountAlreadyExcluded) {
					return nil, fmt.Errorf("failed to exclude account from overall balance: %w", err)
				}
			}
		}
	}

	if err := s.accountRepo.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// FindAccountByID is the use case for finding a account by it's id
func (s *Service) FindAccountByID(ctx context.Context, accountID uuid.UUID) (*Account, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account by id: %w", err)
	}

	return account, nil
}

// FindAccountsByUserID is the use case for finding the users account(s) by the user id
func (s *Service) FindAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error) {
	accounts, err := s.accountRepo.FindAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts by user id: %w", err)
	}

	return accounts, nil
}
