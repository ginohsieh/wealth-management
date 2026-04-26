package store

import (
	"fmt"
	"sync"
)

// Account represents a financial account
type Account struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"` // checking, savings, investment, credit
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Type        string  `json:"type"` // income, expense
}

// Asset represents an investment holding
type Asset struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Quantity      float64 `json:"quantity"`
	PurchasePrice float64 `json:"purchase_price"`
	CurrentPrice  float64 `json:"current_price"`
	Value         float64 `json:"value"`
	GainLoss      float64 `json:"gain_loss"`
	GainLossPct   float64 `json:"gain_loss_pct"`
}

// Summary represents an overall financial snapshot
type Summary struct {
	TotalAssets      float64 `json:"total_assets"`
	TotalLiabilities float64 `json:"total_liabilities"`
	NetWorth         float64 `json:"net_worth"`
	MonthlyIncome    float64 `json:"monthly_income"`
	MonthlyExpenses  float64 `json:"monthly_expenses"`
	MonthlySavings   float64 `json:"monthly_savings"`
	PortfolioValue   float64 `json:"portfolio_value"`
}

// NetWorthSnapshot records a point-in-time financial summary for a given date.
type NetWorthSnapshot struct {
	ID             string  `json:"id"`
	Date           string  `json:"date"`
	PortfolioValue float64 `json:"portfolio_value"`
	TotalAssets    float64 `json:"total_assets"`
	TotalLiabilities float64 `json:"total_liabilities"`
	NetWorth       float64 `json:"net_worth"`
}

var (
	mu                sync.RWMutex
	accounts          []Account
	transactions      []Transaction
	assets            []Asset
	snapshots         []NetWorthSnapshot
	nextAccountID     = 5
	nextTransactionID = 9
	nextAssetID       = 5
	nextSnapshotID    = 1
)

func init() {
	accounts = []Account{
		{ID: "1", Name: "Chase Checking", Type: "checking", Balance: 5250.00, Currency: "USD"},
		{ID: "2", Name: "Chase Savings", Type: "savings", Balance: 15000.00, Currency: "USD"},
		{ID: "3", Name: "Fidelity Investment", Type: "investment", Balance: 45000.00, Currency: "USD"},
		{ID: "4", Name: "Visa Credit Card", Type: "credit", Balance: -1200.00, Currency: "USD"},
	}

	transactions = []Transaction{
		{ID: "1", AccountID: "1", Date: "2024-04-01", Description: "Salary", Amount: 5000.00, Category: "Income", Type: "income"},
		{ID: "2", AccountID: "1", Date: "2024-04-05", Description: "Rent Payment", Amount: -1500.00, Category: "Housing", Type: "expense"},
		{ID: "3", AccountID: "1", Date: "2024-04-10", Description: "Grocery Store", Amount: -200.00, Category: "Food", Type: "expense"},
		{ID: "4", AccountID: "4", Date: "2024-04-12", Description: "Restaurant Dinner", Amount: -85.00, Category: "Food", Type: "expense"},
		{ID: "5", AccountID: "1", Date: "2024-04-15", Description: "Freelance Payment", Amount: 1200.00, Category: "Income", Type: "income"},
		{ID: "6", AccountID: "4", Date: "2024-04-18", Description: "Online Shopping", Amount: -350.00, Category: "Shopping", Type: "expense"},
		{ID: "7", AccountID: "2", Date: "2024-04-20", Description: "Transfer from Checking", Amount: 1000.00, Category: "Transfer", Type: "income"},
		{ID: "8", AccountID: "3", Date: "2024-04-22", Description: "Stock Purchase AAPL", Amount: -2000.00, Category: "Investment", Type: "expense"},
	}

	assets = []Asset{
		{ID: "1", Symbol: "AAPL", Name: "Apple Inc.", Quantity: 10, PurchasePrice: 150.00, CurrentPrice: 185.00},
		{ID: "2", Symbol: "GOOGL", Name: "Alphabet Inc.", Quantity: 5, PurchasePrice: 2800.00, CurrentPrice: 3100.00},
		{ID: "3", Symbol: "MSFT", Name: "Microsoft Corp.", Quantity: 15, PurchasePrice: 280.00, CurrentPrice: 415.00},
		{ID: "4", Symbol: "BRK.B", Name: "Berkshire Hathaway", Quantity: 20, PurchasePrice: 320.00, CurrentPrice: 358.00},
	}

	for i := range assets {
		assets[i] = calculateAssetFields(assets[i])
	}

	snapshots = []NetWorthSnapshot{
		{ID: "1", Date: "2024-04-20", PortfolioValue: 38200.00, TotalAssets: 62000.00, TotalLiabilities: 900.00, NetWorth: 61100.00},
		{ID: "2", Date: "2024-04-21", PortfolioValue: 39100.00, TotalAssets: 62450.00, TotalLiabilities: 1050.00, NetWorth: 61400.00},
		{ID: "3", Date: "2024-04-22", PortfolioValue: 40500.00, TotalAssets: 63100.00, TotalLiabilities: 1100.00, NetWorth: 62000.00},
		{ID: "4", Date: "2024-04-23", PortfolioValue: 41200.00, TotalAssets: 63800.00, TotalLiabilities: 1150.00, NetWorth: 62650.00},
		{ID: "5", Date: "2024-04-24", PortfolioValue: 42000.00, TotalAssets: 64200.00, TotalLiabilities: 1200.00, NetWorth: 63000.00},
	}
	nextSnapshotID = 6
}

func calculateAssetFields(a Asset) Asset {
	a.Value = a.Quantity * a.CurrentPrice
	a.GainLoss = a.Value - (a.Quantity * a.PurchasePrice)
	if a.PurchasePrice > 0 {
		a.GainLossPct = (a.CurrentPrice - a.PurchasePrice) / a.PurchasePrice * 100
	}
	return a
}

// GetAccounts returns all accounts
func GetAccounts() []Account {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Account, len(accounts))
	copy(result, accounts)
	return result
}

// GetAccountByID returns an account by its ID
func GetAccountByID(id string) (Account, bool) {
	mu.RLock()
	defer mu.RUnlock()
	for _, a := range accounts {
		if a.ID == id {
			return a, true
		}
	}
	return Account{}, false
}

// CreateAccount adds a new account and returns it with a generated ID
func CreateAccount(a Account) Account {
	mu.Lock()
	defer mu.Unlock()
	a.ID = fmt.Sprintf("%d", nextAccountID)
	nextAccountID++
	if a.Currency == "" {
		a.Currency = "USD"
	}
	accounts = append(accounts, a)
	return a
}

// UpdateAccount replaces an existing account
func UpdateAccount(id string, a Account) (Account, bool) {
	mu.Lock()
	defer mu.Unlock()
	for i, acc := range accounts {
		if acc.ID == id {
			a.ID = id
			accounts[i] = a
			return a, true
		}
	}
	return Account{}, false
}

// DeleteAccount removes an account by ID
func DeleteAccount(id string) bool {
	mu.Lock()
	defer mu.Unlock()
	for i, a := range accounts {
		if a.ID == id {
			accounts = append(accounts[:i], accounts[i+1:]...)
			return true
		}
	}
	return false
}

// GetTransactions returns all transactions, optionally filtered by account ID
func GetTransactions(accountID string) []Transaction {
	mu.RLock()
	defer mu.RUnlock()
	if accountID == "" {
		result := make([]Transaction, len(transactions))
		copy(result, transactions)
		return result
	}
	var result []Transaction
	for _, t := range transactions {
		if t.AccountID == accountID {
			result = append(result, t)
		}
	}
	return result
}

// CreateTransaction adds a new transaction and returns it with a generated ID
func CreateTransaction(t Transaction) Transaction {
	mu.Lock()
	defer mu.Unlock()
	t.ID = fmt.Sprintf("%d", nextTransactionID)
	nextTransactionID++
	transactions = append(transactions, t)
	return t
}

// GetAssets returns all portfolio assets
func GetAssets() []Asset {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Asset, len(assets))
	copy(result, assets)
	return result
}

// CreateAsset adds a new asset to the portfolio
func CreateAsset(a Asset) Asset {
	mu.Lock()
	defer mu.Unlock()
	a.ID = fmt.Sprintf("%d", nextAssetID)
	nextAssetID++
	a = calculateAssetFields(a)
	assets = append(assets, a)
	return a
}

// UpdateAsset replaces an existing asset
func UpdateAsset(id string, a Asset) (Asset, bool) {
	mu.Lock()
	defer mu.Unlock()
	for i, asset := range assets {
		if asset.ID == id {
			a.ID = id
			a = calculateAssetFields(a)
			assets[i] = a
			return a, true
		}
	}
	return Asset{}, false
}

// DeleteAsset removes an asset by ID
func DeleteAsset(id string) bool {
	mu.Lock()
	defer mu.Unlock()
	for i, a := range assets {
		if a.ID == id {
			assets = append(assets[:i], assets[i+1:]...)
			return true
		}
	}
	return false
}

// GetSummary returns aggregate financial metrics
func GetSummary() Summary {
	mu.RLock()
	defer mu.RUnlock()

	var totalAssets, totalLiabilities float64
	for _, a := range accounts {
		if a.Balance >= 0 {
			totalAssets += a.Balance
		} else {
			totalLiabilities += -a.Balance
		}
	}

	var portfolioValue float64
	for _, a := range assets {
		portfolioValue += a.Value
	}

	var monthlyIncome, monthlyExpenses float64
	for _, t := range transactions {
		switch t.Type {
		case "income":
			monthlyIncome += t.Amount
		case "expense":
			monthlyExpenses += -t.Amount
		}
	}

	return Summary{
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		NetWorth:         totalAssets - totalLiabilities,
		MonthlyIncome:    monthlyIncome,
		MonthlyExpenses:  monthlyExpenses,
		MonthlySavings:   monthlyIncome - monthlyExpenses,
		PortfolioValue:   portfolioValue,
	}
}

// GetSnapshots returns all recorded net worth snapshots
func GetSnapshots() []NetWorthSnapshot {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]NetWorthSnapshot, len(snapshots))
	copy(result, snapshots)
	return result
}

// RecordSnapshot captures the current financial state for the given date.
// If a snapshot already exists for that date it is replaced.
func RecordSnapshot(date string) NetWorthSnapshot {
	mu.Lock()
	defer mu.Unlock()

	var totalAssets, totalLiabilities float64
	for _, a := range accounts {
		if a.Balance >= 0 {
			totalAssets += a.Balance
		} else {
			totalLiabilities += -a.Balance
		}
	}

	var portfolioValue float64
	for _, a := range assets {
		portfolioValue += a.Value
	}

	snap := NetWorthSnapshot{
		Date:             date,
		PortfolioValue:   portfolioValue,
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		NetWorth:         totalAssets - totalLiabilities,
	}

	// Replace existing snapshot for the same date if present
	for i, s := range snapshots {
		if s.Date == date {
			snap.ID = s.ID
			snapshots[i] = snap
			return snap
		}
	}

	snap.ID = fmt.Sprintf("%d", nextSnapshotID)
	nextSnapshotID++
	snapshots = append(snapshots, snap)
	return snap
}
