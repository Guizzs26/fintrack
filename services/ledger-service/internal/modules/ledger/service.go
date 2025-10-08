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

// BalanceAdjustmentParams holds all the requires data for BalanceAdjustment use case
type BalanceAdjustmentParams struct {
	AccountID  uuid.UUID
	UserID     uuid.UUID
	NewBalance int64
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
	account, err := s.FindAccountByID(ctx, params.UserID, params.AccountID)
	if err != nil {
		return fmt.Errorf("failed to find account to add transaction: %w", err)
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
	account, err := s.FindAccountByID(ctx, params.UserID, params.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account to update: %w", err)
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

// ArchiveAccount is the use case for archive an account (important use case with important business logic contribution)
func (s *Service) ArchiveAccount(ctx context.Context, userID, accountID uuid.UUID) error {
	account, err := s.FindAccountByID(ctx, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to find account to archive: %w", err)
	}

	if err := account.Archive(s.clock); err != nil {
		return fmt.Errorf("failed to archive account: %w", err)
	}

	if err := s.accountRepo.Save(ctx, account); err != nil {
		return fmt.Errorf("failed to save archived account state: %w", err)
	}

	return nil
}

// UnarchiveAccount is the use case for unarchive an archived account (important use case with important business logic contribution)
func (s *Service) UnarchiveAccount(ctx context.Context, userID, accountID uuid.UUID) (*Account, error) {
	account, err := s.FindAccountByID(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}

	if err := account.Unarchive(); err != nil {
		return nil, fmt.Errorf("failed to unarchive the account: %w", err)
	}

	if err := s.accountRepo.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save unarchived account: %w", err)
	}

	return account, nil
}

// AdjustAccountBalance is the use case for adjust the balance of an existing accoutn
func (s *Service) AdjustAccountBalance(ctx context.Context, params BalanceAdjustmentParams) (*Account, error) {
	account, err := s.FindAccountByID(ctx, params.UserID, params.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account for balance adjustment: %w", err)
	}

	if err := account.AdjustBalance(params.NewBalance, s.clock); err != nil {
		return nil, fmt.Errorf("failed to adjust account balance: %w", err)
	}

	if err := s.accountRepo.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to save the adjusted account balance: %w", err)
	}

	return account, nil
}

// FindAccountByID is the use case for finding a account by it's id
func (s *Service) FindAccountByID(ctx context.Context, userID, accountID uuid.UUID) (*Account, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account by id: %w", err)
	}

	if account.UserID != userID {
		return nil, ErrAccountNotFound
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
