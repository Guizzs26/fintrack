package ledger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ AccountRepository = (*PostgresAccountRepository)(nil)

var (
	ErrAccountNotFound = errors.New("account not found in database")
)

// ----- Main struct repository and Querier ----- //

// PostgresAccountRepository is a PostgreSQL implementation of the AccountRepository interface defined by the domain layer
type PostgresAccountRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAccountRepository creates a new PostgresAccountRepository
func NewPostgresAccountRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{pool: pool}
}

// ExecTx executes a function within a database transaction
func (par *PostgresAccountRepository) ExecTx(ctx context.Context, fn func(q *Querier) error) error {
	tx, err := par.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("repository: failed to begin transaction: %w", err)
	}

	q := NewQuerier(tx)

	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("repository: transaction rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: failed to commit transaction: %w", err)
	}

	return nil
}

// Querier returns a new Querier instance that uses the repository's connection pool
func (par *PostgresAccountRepository) Querier() *Querier {
	return NewQuerier(par.pool)
}

// DBQuerier is an interface that is satisfied by both *pgxpool.Pool and pgx.Tx
// This allows our query methods to be used both inside and outside of a transaction
// without any changes to the method signatures
type DBQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

// Querier holds a DBQuerier interface, allowing it to execute SQL queries
type Querier struct {
	db DBQuerier
}

// NewQuerier creates a new Querier
func NewQuerier(db DBQuerier) *Querier {
	return &Querier{db: db}
}

// ----- MODELS ----- //

// accountModel represents the account structure in the database
type accountModel struct {
	ID                      uuid.UUID  `db:"id"`
	UserID                  uuid.UUID  `db:"user_id"`
	Name                    string     `db:"name"`
	IncludeInOverallBalance bool       `db:"include_in_overall_balance"`
	ArchivedAt              *time.Time `db:"archived_at"`
	CreatedAt               time.Time  `db:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at"`
}

// transactionModel represents the transaction structure in the database
type transactionModel struct {
	ID          uuid.UUID       `db:"id"`
	AccountID   uuid.UUID       `db:"account_id"`
	UserID      uuid.UUID       `db:"user_id"`
	CategoryID  *uuid.UUID      `db:"category_id"`
	Type        TransactionType `db:"type"`
	Description string          `db:"description"`
	Observation string          `db:"observation"`
	Amount      int64           `db:"amount_in_cents"`
	DueDate     time.Time       `db:"due_date"`
	PaidAt      *time.Time      `db:"paid_at"`
	Metadata    []byte          `db:"metadata"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

// ----- MAPPERS ----- //

// toAccountPersistence maps a domain Account to its persistence model
func toAccountPersistence(a *Account) *accountModel {
	return &accountModel{
		ID:                      a.ID,
		UserID:                  a.UserID,
		Name:                    a.Name,
		IncludeInOverallBalance: a.IncludeInOverallBalance,
		ArchivedAt:              a.GetArchivedAt(),
	}
}

// toTransactionPersistence maps a domain Transaction to its persistence model
func toTransactionPersistence(tx *Transaction, accountID, userID uuid.UUID) *transactionModel {
	return &transactionModel{
		ID:          tx.ID,
		AccountID:   accountID,
		UserID:      userID,
		CategoryID:  tx.CategoryID,
		Type:        tx.Type,
		Description: tx.Description,
		Observation: tx.Observation,
		Amount:      tx.Amount,
		DueDate:     tx.DueDate,
		PaidAt:      tx.PaidAt,
		Metadata:    nil,
	}
}

// toAccountDomain maps a persistence accountModel and its transactions to a domain Account
func toAccountDomain(m *accountModel, txsModels []transactionModel) *Account {
	domainTx := make([]Transaction, len(txsModels))
	for i, txm := range txsModels {
		domainTx[i] = *toTransactionDomain(&txm)
	}

	// 2. Montar o agregado Account, injetando suas transações filhas.
	// Note que não usamos NewAccount() aqui, pois estamos recriando um agregado
	// que já existe, e não criando um novo.
	return &Account{
		ID:           m.ID,
		UserID:       m.UserID,
		Name:         m.Name,
		ArchivedAt:   m.ArchivedAt,
		transactions: domainTx,
	}
}

// toTransactionDomain maps a persistence transactionModel to a domain Transaction
func toTransactionDomain(m *transactionModel) *Transaction {
	return &Transaction{
		ID:          m.ID,
		Type:        m.Type,
		Description: m.Description,
		Observation: m.Observation,
		Amount:      m.Amount,
		DueDate:     m.DueDate,
		PaidAt:      m.PaidAt,
	}
}

// ----- Repository Methods ----- //

// Save persists the entire Account aggregate. It operates transactionally,
// first upserting the account, then deleting all existing transactions for that account,
// and finally bulk-inserting the current transactions from the aggregate
func (par *PostgresAccountRepository) Save(ctx context.Context, account *Account) error {
	return par.ExecTx(ctx, func(q *Querier) error {
		accModel := toAccountPersistence(account)

		if err := q.upsertAccount(ctx, accModel); err != nil {
			return err
		}

		if err := q.deleteTransactionsForAccount(ctx, accModel.ID); err != nil {
			return err
		}

		if err := q.bulkInsertTransactions(ctx, account.ID, account.UserID, account.Transactions()); err != nil {
			return err
		}

		return nil
	})
}

// FindByID retrieves an Account aggregate by its ID. It first fetches the account
// and then all its associated transactions, reconstructing the full domain aggregate
func (par *PostgresAccountRepository) FindByID(ctx context.Context, accountID uuid.UUID) (*Account, error) {
	q := par.Querier()

	accModel, err := q.getAccountByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	txModels, err := q.getTransactionsByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions for account: %w", err)
	}

	account := toAccountDomain(accModel, txModels)
	return account, nil
}

// ----- Querier Methods ----- //

// upsertAccount inserts a new account or updates an existing one based on its ID
// It uses the 'ON CONFLICT' clause to perform an update if the account already exists
func (q *Querier) upsertAccount(ctx context.Context, accountModel *accountModel) error {
	query := `
		INSERT INTO accounts (
			id, 
			user_id, 
			name, 
			include_in_overall_balance, 
			archived_at
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id)
		DO UPDATE SET 
			name = EXCLUDED.name,
   	  include_in_overall_balance = EXCLUDED.include_in_overall_balance,
    	archived_at = EXCLUDED.archived_at,
    	updated_at = now()
	`

	_, err := q.db.Exec(ctx, query,
		accountModel.ID,
		accountModel.UserID,
		accountModel.Name,
		accountModel.IncludeInOverallBalance,
		accountModel.ArchivedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert account: %v", err)
	}

	return nil
}

// deleteTransactionsForAccount deletes all transactions associated with a given account ID
func (q *Querier) deleteTransactionsForAccount(ctx context.Context, accountID uuid.UUID) error {
	query := `DELETE FROM transactions WHERE account_id = $1`

	_, err := q.db.Exec(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("failed to delete transactions for account: %v", err)
	}

	return nil
}

// bulkInsertTransactions efficiently inserts a slice of transactions in a single batch operation
func (q *Querier) bulkInsertTransactions(ctx context.Context, accountID, userID uuid.UUID, transactions []Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO transactions (id, account_id, user_id, category_id, type, description, observation, amount_in_cents, due_date, paid_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, tx := range transactions {
		txModel := toTransactionPersistence(&tx, accountID, userID)
		batch.Queue(query,
			txModel.ID,
			txModel.AccountID,
			txModel.UserID,
			txModel.CategoryID,
			txModel.Type,
			txModel.Description,
			txModel.Observation,
			txModel.Amount,
			txModel.DueDate,
			txModel.PaidAt,
		)
	}

	br := q.db.SendBatch(ctx, batch)
	defer br.Close()

	if _, err := br.Exec(); err != nil {
		return fmt.Errorf("failed to bulk insert transactions: %v", err)
	}

	return nil
}

// getAccountByID retrieves a single account from the database by its ID
func (q *Querier) getAccountByID(ctx context.Context, accountID uuid.UUID) (*accountModel, error) {
	query := `
		SELECT id, user_id, name, include_in_overall_balance, archived_at, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var m accountModel
	err := q.db.QueryRow(ctx, query, accountID).Scan(
		&m.ID,
		&m.UserID,
		&m.Name,
		&m.IncludeInOverallBalance,
		&m.ArchivedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to fetch account by id: %w", err)
	}

	return &m, nil
}

// getTransactionsByAccountID retrieves all transactions for a given account ID
func (q *Querier) getTransactionsByAccountID(ctx context.Context, accountID uuid.UUID) ([]transactionModel, error) {
	query := `
		SELECT id, account_id, user_id, category_id, type, description, 
			observation, amount_in_cents, due_date, metadata, paid_at,
			created_at, updated_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY due_date ASC
	`

	rows, err := q.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions for account: %v", err)
	}
	defer rows.Close()

	var transactions []transactionModel
	for rows.Next() {
		var m transactionModel
		if err := rows.Scan(
			&m.ID,
			&m.AccountID,
			&m.UserID,
			&m.CategoryID,
			&m.Type,
			&m.Description,
			&m.Observation,
			&m.Amount,
			&m.DueDate,
			&m.PaidAt,
			&m.Metadata,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %v", err)
		}
		transactions = append(transactions, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed during transaction rows iteration: %v", err)
	}

	return transactions, nil
}
