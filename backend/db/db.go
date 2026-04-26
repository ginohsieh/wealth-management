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
    currency TEXT          NOT NULL DEFAULT 'USD'
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

CREATE TABLE IF NOT EXISTS assets (
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
    enabled          BOOLEAN NOT NULL DEFAULT FALSE
);
`

// seedIfEmpty inserts demo data only when the accounts table is empty.
func seedIfEmpty(database *sql.DB) error {
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	log.Println("[db] seeding initial data")

	if _, err := database.Exec(`
		INSERT INTO accounts(name, type, balance, currency) VALUES
		('Chase Checking',      'checking',   5250.00, 'USD'),
		('Chase Savings',       'savings',   15000.00, 'USD'),
		('Fidelity Investment', 'investment', 45000.00, 'USD'),
		('Visa Credit Card',    'credit',    -1200.00, 'USD')`); err != nil {
		return err
	}

	if _, err := database.Exec(`
		INSERT INTO transactions(account_id, date, description, amount, category, type) VALUES
		(1, '2024-04-01', 'Salary',                    5000.00, 'Income',     'income'),
		(1, '2024-04-05', 'Rent Payment',             -1500.00, 'Housing',    'expense'),
		(1, '2024-04-10', 'Grocery Store',             -200.00, 'Food',       'expense'),
		(4, '2024-04-12', 'Restaurant Dinner',          -85.00, 'Food',       'expense'),
		(1, '2024-04-15', 'Freelance Payment',         1200.00, 'Income',     'income'),
		(4, '2024-04-18', 'Online Shopping',           -350.00, 'Shopping',   'expense'),
		(2, '2024-04-20', 'Transfer from Checking',    1000.00, 'Transfer',   'income'),
		(3, '2024-04-22', 'Stock Purchase AAPL',      -2000.00, 'Investment', 'expense')`); err != nil {
		return err
	}

	if _, err := database.Exec(`
		INSERT INTO assets(symbol, name, quantity, purchase_price, current_price) VALUES
		('AAPL',  'Apple Inc.',          10, 150.00,  185.00),
		('GOOGL', 'Alphabet Inc.',        5, 2800.00, 3100.00),
		('MSFT',  'Microsoft Corp.',     15, 280.00,   415.00),
		('BRK.B', 'Berkshire Hathaway',  20, 320.00,   358.00)`); err != nil {
		return err
	}

	if _, err := database.Exec(`
		INSERT INTO portfolio_trades(symbol, name, type, date, quantity, price, market, fee, tax, notes) VALUES
		('AAPL',  'Apple Inc.',          'buy',  '2023-11-15', 10,  150.00, 'US', 0, 0, 'Initial position'),
		('GOOGL', 'Alphabet Inc.',       'buy',  '2023-12-01',  5, 2800.00, 'US', 0, 0, ''),
		('MSFT',  'Microsoft Corp.',     'buy',  '2024-01-10', 15,  280.00, 'US', 0, 0, ''),
		('BRK.B', 'Berkshire Hathaway',  'buy',  '2024-02-20', 20,  320.00, 'US', 0, 0, ''),
		('AAPL',  'Apple Inc.',          'buy',  '2024-03-05',  5,  170.00, 'US', 0, 0, 'Adding to position'),
		('MSFT',  'Microsoft Corp.',     'sell', '2024-04-10',  3,  400.00, 'US', 0, 0, 'Partial sell'),
		('GOOGL', 'Alphabet Inc.',       'buy',  '2024-04-15',  2, 3000.00, 'US', 0, 0, ''),
		('BRK.B', 'Berkshire Hathaway',  'sell', '2024-04-20',  5,  350.00, 'US', 0, 0, '')`); err != nil {
		return err
	}

	if _, err := database.Exec(`
		INSERT INTO symbol_prices(symbol, price) VALUES
		('AAPL',  185.00),
		('GOOGL', 3100.00),
		('MSFT',   415.00),
		('BRK.B',  358.00)
		ON CONFLICT DO NOTHING`); err != nil {
		return err
	}

	if _, err := database.Exec(`
		INSERT INTO net_worth_snapshots(date, portfolio_value, total_assets, total_liabilities, net_worth) VALUES
		('2024-04-20', 38200.00, 62000.00,  900.00, 61100.00),
		('2024-04-21', 39100.00, 62450.00, 1050.00, 61400.00),
		('2024-04-22', 40500.00, 63100.00, 1100.00, 62000.00),
		('2024-04-23', 41200.00, 63800.00, 1150.00, 62650.00),
		('2024-04-24', 42000.00, 64200.00, 1200.00, 63000.00)`); err != nil {
		return err
	}

	if _, err := database.Exec(`
		INSERT INTO price_config(id, source, interval_seconds, api_key, enabled)
		VALUES(1, 'yahoo', 300, '', FALSE)
		ON CONFLICT DO NOTHING`); err != nil {
		return err
	}

	return nil
}
