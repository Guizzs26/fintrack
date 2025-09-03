package ledger

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrAccountArchived          = errors.New("account is already archived")
	ErrAccountNotArchived       = errors.New("account is not archived")
	ErrTransactionNotFound      = errors.New("transaction not found in this account")
	ErrTransactionAlreadyPaid   = errors.New("transaction is already marked as paid")
	ErrTransactionAlreadyUnpaid = errors.New("transaction is already marked as unpaid")
	ErrPaymentDateInFuture      = errors.New("payment date cannot be in the future")
	ErrAmountCannotBeZero       = errors.New("transaction amount cannot be zero")
	ErrDescriptionRequired      = errors.New("transaction description is required")
	ErrAccountNameRequired      = errors.New("account name is required")
	ErrInconsistentAmountSign   = errors.New("transaction amount sign is inconsistent with its type")
	ErrInvalidTransactionType   = errors.New("invalid transaction type")
)

const (
	Income     TransactionType = "INCOME"
	Expense    TransactionType = "EXPENSE"
	Adjustment TransactionType = "ADJUSTMENT"

	maxAccountNameLength            = 100
	maxTransactionDescriptionLength = 100
	maxTransactionObservationLength = 2500
)

// TransactionType represents the type of a financial transaction
type TransactionType string

// Transaction represents a single financial entry in an account
type Transaction struct {
	ID          uuid.UUID
	Type        TransactionType
	Description string
	Observation string
	Amount      int64
	DueDate     time.Time
	PaidAt      *time.Time
}

// Account represents a user's account, which holds a collection of transactions (our aggregate root)
type Account struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         string
	transactions []Transaction
	ArchivedAt   *time.Time
}

// NewAccount creates a new Account with the given user ID and name
// It returns an error if the name is empty or exceeds the maximum length
func NewAccount(userID uuid.UUID, name string) (*Account, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrAccountNameRequired
	}
	if len(name) > maxAccountNameLength {
		return nil, fmt.Errorf("account name cannot exceed %d characters", maxAccountNameLength)
	}

	return &Account{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         name,
		transactions: make([]Transaction, 0),
	}, nil
}

// AddTransaction adds a new transaction to the account
// It returns an error if the transaction amount is zero
func (a *Account) AddTransaction(txType TransactionType, description, observation string, amount int64, dueDate time.Time, paidAt *time.Time) error {
	if a.ArchivedAt != nil {
		return ErrAccountArchived
	}

	if strings.TrimSpace(description) == "" {
		return ErrDescriptionRequired
	}
	if len(description) > maxTransactionDescriptionLength {
		return fmt.Errorf("transaction description cannot exceed %d characters", maxTransactionDescriptionLength)
	}

	if strings.TrimSpace(observation) != "" {
		if utf8.RuneCountInString(observation) > maxTransactionObservationLength {
			return fmt.Errorf("transaction observation cannot exceed %d characters", maxTransactionObservationLength)
		}
	}

	if amount == 0 {
		return ErrAmountCannotBeZero
	}

	switch txType {
	case Income, Expense, Adjustment:
		// valid type
	default:
		return ErrInvalidTransactionType
	}

	isIncome := txType == Income
	isExpense := txType == Expense
	if (isIncome && amount < 0) || (isExpense && amount > 0) {
		return ErrInconsistentAmountSign
	}

	if paidAt != nil && paidAt.After(time.Now()) {
		return ErrPaymentDateInFuture
	}

	tx := Transaction{
		ID:          uuid.New(),
		Type:        txType,
		Amount:      amount,
		Description: description,
		Observation: observation,
		DueDate:     dueDate,
		PaidAt:      paidAt,
	}

	a.transactions = append(a.transactions, tx)

	return nil
}

func (a *Account) DeleteTransaction(txID uuid.UUID) error {
	if a.ArchivedAt != nil {
		return ErrAccountArchived
	}

	foundIndex := -1
	for i, tx := range a.transactions {
		if tx.ID == txID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return ErrTransactionNotFound
	}

	a.transactions = append(a.transactions[:foundIndex], a.transactions[foundIndex+1:]...)

	return nil
}

func (a *Account) Balance() int64 {
	var total int64 = 0
	for _, tx := range a.transactions {
		total += tx.Amount
	}
	return total
}

func (a *Account) MarkTransactionAsPaid(txID uuid.UUID, paidAt time.Time) error {
	if a.ArchivedAt != nil {
		return ErrAccountArchived
	}

	if paidAt.After(time.Now()) {
		return ErrPaymentDateInFuture
	}

	target, err := a.findTransaction(txID)
	if err != nil {
		return err
	}

	if target.PaidAt != nil {
		return ErrTransactionAlreadyPaid
	}

	target.PaidAt = &paidAt

	return nil
}

func (a *Account) MarkTransactionAsUnpaid(txID uuid.UUID) error {
	if a.ArchivedAt != nil {
		return ErrAccountArchived
	}

	target, err := a.findTransaction(txID)
	if err != nil {
		return err
	}

	if target.PaidAt == nil {
		return ErrTransactionAlreadyUnpaid
	}

	target.PaidAt = nil

	return nil
}

func (a *Account) Archive() error {
	if a.ArchivedAt != nil {
		return ErrAccountArchived
	}

	now := time.Now()
	a.ArchivedAt = &now

	return nil
}

func (a *Account) Unarchive() error {
	if a.ArchivedAt == nil {
		return ErrAccountNotArchived
	}

	a.ArchivedAt = nil

	return nil
}

func (a *Account) findTransaction(txID uuid.UUID) (*Transaction, error) {
	for i := range a.transactions {
		if txID == a.transactions[i].ID {
			return &a.transactions[i], nil
		}
	}
	return nil, ErrTransactionNotFound
}
