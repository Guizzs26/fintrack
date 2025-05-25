package postgres

import (
	"context"
	"database/sql"

	"github.com/Guizzs26/fintrack/internal/modules/identity/auth/infra/persistence"
)

type PostgresAuthRepository struct {
	db *sql.DB
}

func NewPostgresAuthRepository(db *sql.DB) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db}
}

func (r *PostgresAuthRepository) Create(ctx context.Context, auth *persistence.AuthDB) error {
	query := `INSERT INTO auth (id, email, password, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, auth.ID, auth.Email, auth.Password, auth.CreatedAt)
	return err
}

func (r *PostgresAuthRepository) FindByEmail(ctx context.Context, email string) (*persistence.AuthDB, error) {
	query := `SELECT id, email, password, created_at FROM auth WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	var auth persistence.AuthDB
	err := row.Scan(&auth.ID, &auth.Email, &auth.Password, &auth.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &auth, nil
}
