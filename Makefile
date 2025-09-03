# Makefile for Goose DB Migrations

-include .env
export

## db/new: Create a new SQL migration file. Ex: make db/new name=create_users_table
db/new:
	@echo ">> Creating new migration: $(name)"
	@goose create $(name) sql

## db/status: Show the status of all migrations.
db/status:
	@echo ">> Checking migration status..."
	@goose status

## db/up: Apply all pending migrations.
db/up:
	@echo ">> Applying all pending migrations..."
	@goose up

## db/down: Roll back the last applied migration.
db/down:
	@echo ">> Reverting last migration..."
	@goose down

## db/redo: Roll back and reapply the last migration. Useful for testing a migration.
db/redo:
	@echo ">> Redoing last migration..."
	@goose redo

# --- Advanced Commands ---

## db/up-by-one: Apply the next pending migration.
db/up-by-one:
	@echo ">> Applying next migration..."
	@goose up-by-one

## db/up-to: Migrate up to a specific version. Ex: make db/up-to version=2025...
db/up-to:
	@echo ">> Migrating up to version $(version)..."
	@goose up-to $(version)

## db/down-to: Roll back to a specific version. Ex: make db/down-to version=2025...
db/down-to:
	@echo ">> Rolling back to version $(version)..."
	@goose down-to $(version)

## db/reset: Roll back ALL migrations. WARNING: THIS DELETES DATA.
db/reset:
	@echo ">> WARNING: Rolling back all migrations..."
	@goose reset

## db/version: Show the current schema version in the database.
db/version:
	@echo ">> Current database version:"
	@goose version

## db/validate: Check migration file syntax and order without applying them.
db/validate:
	@echo ">> Validating migration files..."
	@goose validate

## db/fix: Convert timestamp-based migration filenames to sequential.
db/fix:
	@echo ">> Fixing migration versioning to sequential..."
	@goose fix

# --- Help ---

## help: Show this help message.
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -h -E '^##' $(MAKEFILE_LIST) | \
	awk 'match($$0, /^## ([^:]+): (.*)/, a) { \
		printf "  \033[36m%-15s\033[0m %s\n", a[1], a[2] \
	}'


.PHONY: db/new db/status db/up db/down db/redo db/up-by-one db/up-to db/down-to db/reset db/version db/validate db/fix help