-- +goose Up
-- +goose StatementBegin

CREATE TYPE transaction_type AS ENUM (
  'INCOME',
  'EXPENSE',
  'ADJUSTMENT'
);

CREATE TABLE IF NOT EXISTS transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL,
  user_id UUID NOT NULL, 
  category_id UUID,
  type transaction_type NOT NULL,
  description VARCHAR(100) NOT NULL,
  observation TEXT,
  amount_in_cents BIGINT NOT NULL,
  due_date TIMESTAMPTZ NOT NULL,
  paid_at TIMESTAMPTZ,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT fk_accounts FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_categories FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions (account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id_paid_at ON transactions (user_id, paid_at);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id_due_date ON transactions (user_id, due_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_account_id;
DROP INDEX IF EXISTS idx_transactions_user_id_paid_at;
DROP INDEX IF EXISTS idx_transactions_user_id_due_date;
DROP TABLE IF EXISTS transactions;
DROP TYPE IF EXISTS transaction_type
-- +goose StatementEnd
