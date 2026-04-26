// Package db manages the PostgreSQL connection and schema lifecycle.
//
// Connection settings are read from environment variables:
//
//	DATABASE_URL  – full DSN (takes precedence over individual vars)
//	DB_HOST       – default: localhost
//	DB_PORT       – default: 5432
//	DB_NAME       – default: wealthdb
//	DB_USER       – default: wealthuser
//	DB_PASSWORD   – default: wealthpass
//	DB_SSLMODE    – default: disable
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Open connects to PostgreSQL and verifies the connection with a ping.
func Open() (*sql.DB, error) {
	database, err := sql.Open("postgres", DSN())
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return database, nil
}

// DSN builds the PostgreSQL connection string from environment variables.
// DATABASE_URL takes precedence; otherwise individual DB_* vars are used.
func DSN() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		envOr("DB_HOST", "localhost"),
		envOr("DB_PORT", "5432"),
		envOr("DB_NAME", "wealthdb"),
		envOr("DB_USER", "wealthuser"),
		envOr("DB_PASSWORD", "wealthpass"),
		envOr("DB_SSLMODE", "disable"),
	)
}

// Migrate creates tables (if not already present) and seeds initial data when
// the database is empty.
func Migrate(database *sql.DB) error {
	if err := createSchema(database); err != nil {
		return fmt.Errorf("db: schema: %w", err)
	}
	if err := seedIfEmpty(database); err != nil {
		return fmt.Errorf("db: seed: %w", err)
	}
	log.Println("[db] schema ready")
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// createSchema runs idempotent DDL to ensure all tables exist.
func createSchema(database *sql.DB) error {
	_, err := database.Exec(ddl)
	return err
}

const ddl = `
CREATE TABLE IF NOT EXISTS accounts (
    id       SERIAL        PRIMARY KEY,
    name     TEXT          NOT NULL,
    type     TEXT          NOT NULL,
    balance  NUMERIC(18,4) NOT NULL DEFAULT 0,
    currency TEXT          NOT NULL DEFAULT 'TWD'
);

CREATE TABLE IF NOT EXISTS transactions (
    id          SERIAL        PRIMARY KEY,
    account_id  INTEGER       NOT NULL,
    date        TEXT          NOT NULL,
    description TEXT          NOT NULL,
    amount      NUMERIC(18,4) NOT NULL,
    category    TEXT          NOT NULL,
    type        TEXT          NOT NULL
);

-- Rename legacy flat-model assets table if it still has old schema columns.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'assets' AND column_name = 'purchase_price'
    ) THEN
        ALTER TABLE assets RENAME TO portfolio_positions;
    END IF;
END $$;

-- New assets table: bridge between accounts and transactions.
-- type: 'cash' | 'stock' | 'credit_line'
CREATE TABLE IF NOT EXISTS assets (
    id         SERIAL  PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    type       TEXT    NOT NULL,
    name       TEXT    NOT NULL,
    symbol     TEXT,
    currency   TEXT    NOT NULL DEFAULT 'TWD'
);

-- Per-asset daily net-value snapshot.
CREATE TABLE IF NOT EXISTS asset_daily_values (
    id       SERIAL        PRIMARY KEY,
    asset_id INTEGER       NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    date     DATE          NOT NULL,
    value    NUMERIC(18,4) NOT NULL,
    currency TEXT          NOT NULL DEFAULT 'TWD',
    UNIQUE (asset_id, date)
);

-- Add asset_id link to transactions (nullable for backward compatibility).
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS asset_id INTEGER REFERENCES assets(id) ON DELETE SET NULL;

-- Legacy portfolio_positions table (renamed from old assets).
-- Created only when the DO $$ block above could not rename (fresh DB).
CREATE TABLE IF NOT EXISTS portfolio_positions (
    id             SERIAL        PRIMARY KEY,
    symbol         TEXT          NOT NULL,
    name           TEXT          NOT NULL,
    quantity       NUMERIC(18,8) NOT NULL,
    purchase_price NUMERIC(18,4) NOT NULL,
    current_price  NUMERIC(18,4) NOT NULL
);

CREATE TABLE IF NOT EXISTS portfolio_trades (
    id       SERIAL        PRIMARY KEY,
    symbol   TEXT          NOT NULL,
    name     TEXT          NOT NULL,
    type     TEXT          NOT NULL,
    date     TEXT          NOT NULL,
    quantity NUMERIC(18,8) NOT NULL,
    price    NUMERIC(18,4) NOT NULL,
    market   TEXT          NOT NULL DEFAULT 'US',
    fee      NUMERIC(18,4) NOT NULL DEFAULT 0,
    tax      NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes    TEXT          NOT NULL DEFAULT ''
);

ALTER TABLE portfolio_trades ADD COLUMN IF NOT EXISTS account_id INTEGER REFERENCES accounts(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS symbol_prices (
    symbol       TEXT          PRIMARY KEY,
    price        NUMERIC(18,4) NOT NULL,
    last_updated TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS net_worth_snapshots (
    id                SERIAL        PRIMARY KEY,
    date              TEXT          UNIQUE NOT NULL,
    portfolio_value   NUMERIC(18,4) NOT NULL,
    total_assets      NUMERIC(18,4) NOT NULL,
    total_liabilities NUMERIC(18,4) NOT NULL,
    net_worth         NUMERIC(18,4) NOT NULL
);

CREATE TABLE IF NOT EXISTS price_config (
    id               INTEGER PRIMARY KEY DEFAULT 1,
    source           TEXT    NOT NULL DEFAULT 'yahoo',
    interval_seconds INTEGER NOT NULL DEFAULT 300,
    api_key          TEXT    NOT NULL DEFAULT '',
    enabled          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS symbol_names (
    symbol     TEXT        PRIMARY KEY,
    name       TEXT        NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS account_groups (
    id          SERIAL PRIMARY KEY,
    name        TEXT   NOT NULL,
    description TEXT   NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS account_group_members (
    group_id   INTEGER NOT NULL REFERENCES account_groups(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, account_id)
);
`

// seedIfEmpty inserts initial config only when the accounts table is empty.
func seedIfEmpty(database *sql.DB) error {
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	log.Println("[db] seeding initial data")

	if _, err := database.Exec(`
		INSERT INTO price_config(id, source, interval_seconds, api_key, enabled)
		VALUES(1, 'yahoo', 300, '', TRUE)
		ON CONFLICT DO NOTHING`); err != nil {
		return err
	}

	return nil
}
