package ledger

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxAccountNameLength            = 100
	maxTransactionDescriptionLength = 100
	maxTransactionObservationLength = 2500
)

// TransactionType represents the type of a financial transaction
type TransactionType string

const (
	Income     TransactionType = "INCOME"
	Expense    TransactionType = "EXPENSE"
	Adjustment TransactionType = "ADJUSTMENT"
)

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
		return nil, errors.New("account name cannot be empty")
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
		return errors.New("cannot add a transaction to an archived account")
	}

	if strings.TrimSpace(description) == "" {
		return errors.New("transaction description cannot be empty")
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
		return errors.New("transaction amount cannot be zero")
	}

	switch txType {
	case Income, Expense, Adjustment:
		// valid type
	default:
		return fmt.Errorf("invalid transaction type: %s", txType)
	}

	isIncome := txType == Income
	isExpense := txType == Expense
	if (isIncome && amount < 0) || (isExpense && amount > 0) {
		return fmt.Errorf("transaction amount sign is inconsistent: got %d for type %s", amount, txType)
	}

	if paidAt != nil && paidAt.After(time.Now()) {
		return errors.New("transaction payment date cannot be in the future")
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
		return errors.New("cannot modify an archived account")
	}

	foundIndex := -1
	for i, tx := range a.transactions {
		if tx.ID == txID {
			fmt.Println("Verificando o index:", i)
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return errors.New("transaction not found in this account")
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
		return errors.New("cannot modify an archived account")
	}

	if paidAt.After(time.Now()) {
		return errors.New("payment date cannot be in the future")
	}

	var target *Transaction
	found := false
	for i := range a.transactions {
		if a.transactions[i].ID == txID {
			target = &a.transactions[i]
			found = true
			break
		}
	}

	if !found {
		return errors.New("transaction not found in this account")
	}

	if target.PaidAt != nil {
		return errors.New("transaction is already marked as paid")
	}

	target.PaidAt = &paidAt

	return nil
}

func (a *Account) MarkTransactionAsUnpaid(txID uuid.UUID) error {
	if a.ArchivedAt != nil {
		return errors.New("cannot modify an archived account")
	}

	var target *Transaction
	found := false
	for i := range a.transactions {
		if a.transactions[i].ID == txID {
			target = &a.transactions[i]
			found = true
			break
		}
	}

	if !found {
		return errors.New("transaction not found in this account")
	}

	if target.PaidAt == nil {
		return errors.New("transaction is already marked as unpaid")
	}

	target.PaidAt = nil

	return nil
}

func (a *Account) Archive() error {
	if a.ArchivedAt != nil {
		return errors.New("account is already archived")
	}

	now := time.Now()
	a.ArchivedAt = &now

	return nil
}

func (a *Account) Unarchive() error {
	if a.ArchivedAt == nil {
		return errors.New("account is not archived")
	}

	a.ArchivedAt = nil

	return nil
}
