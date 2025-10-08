-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID, -- Can be NULL for system-default categories
  parent_id UUID, -- For creating sub-categories
  name VARCHAR(100) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT fk_users
    FOREIGN KEY(user_id) 
    REFERENCES users(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_parent_category
    FOREIGN KEY(parent_id) 
    REFERENCES categories(id)
    ON DELETE CASCADE
);

-- Create an index on user_id to quickly find user-defined categories
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_categories_user_id;
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
