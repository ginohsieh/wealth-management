package store

import (
	"fmt"
	"sync"
	"time"
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

// Asset represents an investment holding (legacy flat model, kept for snapshot compatibility)
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

// PortfolioTrade records a single buy or sell event for a security.
// Fee and Tax are auto-calculated from market rules when not explicitly provided.
type PortfolioTrade struct {
	ID       string  `json:"id"`
	Symbol   string  `json:"symbol"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`     // "buy" or "sell"
	Date     string  `json:"date"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Market   string  `json:"market"` // "US", "TW", "HK", "OTHER"
	Fee      float64 `json:"fee"`
	Tax      float64 `json:"tax"`
	Notes    string  `json:"notes"`
}

// Holding is a derived, aggregated position computed from all trades for a symbol.
type Holding struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Quantity     float64 `json:"quantity"`      // net shares held
	AvgCost      float64 `json:"avg_cost"`      // average cost per share (incl. fees)
	CurrentPrice float64 `json:"current_price"` // latest market price
	Value        float64 `json:"value"`         // quantity * current_price
	TotalCost    float64 `json:"total_cost"`    // quantity * avg_cost
	GainLoss     float64 `json:"gain_loss"`     // value - total_cost
	GainLossPct  float64 `json:"gain_loss_pct"` // gain_loss / total_cost * 100
	LastUpdated  string  `json:"last_updated"`  // RFC3339 timestamp of last price fetch
}

// PriceConfig holds the configurable settings for the background price poller.
type PriceConfig struct {
	// Source selects the price provider: "yahoo" or "alphavantage"
	Source          string `json:"source"`
	// IntervalSeconds is how often to poll (minimum 60, default 300)
	IntervalSeconds int    `json:"interval_seconds"`
	// APIKey is required for Alpha Vantage; ignored for Yahoo
	APIKey          string `json:"api_key"`
	// Enabled turns polling on or off without changing other settings
	Enabled         bool   `json:"enabled"`
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
	mu                  sync.RWMutex
	accounts            []Account
	transactions        []Transaction
	assets              []Asset
	trades              []PortfolioTrade
	symbolPrices        map[string]float64 // current market price per symbol
	symbolLastUpdated   map[string]time.Time // when each symbol's price was last fetched
	priceConfig         PriceConfig
	snapshots           []NetWorthSnapshot
	nextAccountID       = 5
	nextTransactionID   = 9
	nextAssetID         = 5
	nextTradeID         = 9
	nextSnapshotID      = 1
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

	// Seed portfolio trades (transaction-based portfolio)
	symbolPrices = map[string]float64{
		"AAPL":  185.00,
		"GOOGL": 3100.00,
		"MSFT":  415.00,
		"BRK.B": 358.00,
	}
	symbolLastUpdated = map[string]time.Time{}
	priceConfig = PriceConfig{
		Source:          "yahoo",
		IntervalSeconds: 300,
		Enabled:         false,
	}
	trades = []PortfolioTrade{
		{ID: "1", Symbol: "AAPL", Name: "Apple Inc.", Type: "buy", Date: "2023-11-15", Quantity: 10, Price: 150.00, Market: "US", Fee: 0, Tax: 0, Notes: "Initial position"},
		{ID: "2", Symbol: "GOOGL", Name: "Alphabet Inc.", Type: "buy", Date: "2023-12-01", Quantity: 5, Price: 2800.00, Market: "US", Fee: 0, Tax: 0, Notes: ""},
		{ID: "3", Symbol: "MSFT", Name: "Microsoft Corp.", Type: "buy", Date: "2024-01-10", Quantity: 15, Price: 280.00, Market: "US", Fee: 0, Tax: 0, Notes: ""},
		{ID: "4", Symbol: "BRK.B", Name: "Berkshire Hathaway", Type: "buy", Date: "2024-02-20", Quantity: 20, Price: 320.00, Market: "US", Fee: 0, Tax: 0, Notes: ""},
		{ID: "5", Symbol: "AAPL", Name: "Apple Inc.", Type: "buy", Date: "2024-03-05", Quantity: 5, Price: 170.00, Market: "US", Fee: 0, Tax: 0, Notes: "Adding to position"},
		{ID: "6", Symbol: "MSFT", Name: "Microsoft Corp.", Type: "sell", Date: "2024-04-10", Quantity: 3, Price: 400.00, Market: "US", Fee: 0, Tax: 0, Notes: "Partial sell"},
		{ID: "7", Symbol: "GOOGL", Name: "Alphabet Inc.", Type: "buy", Date: "2024-04-15", Quantity: 2, Price: 3000.00, Market: "US", Fee: 0, Tax: 0, Notes: ""},
		{ID: "8", Symbol: "BRK.B", Name: "Berkshire Hathaway", Type: "sell", Date: "2024-04-20", Quantity: 5, Price: 350.00, Market: "US", Fee: 0, Tax: 0, Notes: ""},
	}
	nextTradeID = 9

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

// CalcFeeAndTax returns the broker fee and transaction tax for a trade
// based on market rules. Values are in the same currency as the trade.
//
// Rules applied:
//   US  – fee=0 (commission-free brokers), tax=0
//   TW  – fee=0.1425% of trade value (broker commission, both sides),
//          tax=0.3% of sell value (securities transaction tax, sell only)
//   HK  – fee=0.25% of trade value (broker commission, both sides),
//          stamp duty=0.1% of trade value (both sides)
//   OTHER/default – fee=0, tax=0
func CalcFeeAndTax(market, tradeType string, tradeValue float64) (fee, tax float64) {
	switch market {
	case "TW":
		fee = tradeValue * 0.001425
		if tradeType == "sell" {
			tax = tradeValue * 0.003
		}
	case "HK":
		fee = tradeValue * 0.0025
		tax = tradeValue * 0.001
	default: // "US", "OTHER", etc.
		fee = 0
		tax = 0
	}
	return
}

// GetTrades returns all portfolio trades
func GetTrades() []PortfolioTrade {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]PortfolioTrade, len(trades))
	copy(result, trades)
	return result
}

// CreateTrade records a new buy or sell trade.
// Fee and Tax are auto-calculated when the caller leaves them as 0.
func CreateTrade(t PortfolioTrade) PortfolioTrade {
	mu.Lock()
	defer mu.Unlock()

	t.ID = fmt.Sprintf("%d", nextTradeID)
	nextTradeID++

	tradeValue := t.Quantity * t.Price
	autoFee, autoTax := CalcFeeAndTax(t.Market, t.Type, tradeValue)
	if t.Fee == 0 {
		t.Fee = autoFee
	}
	if t.Tax == 0 {
		t.Tax = autoTax
	}

	trades = append(trades, t)

	// Initialise current price for new symbols
	if _, exists := symbolPrices[t.Symbol]; !exists {
		symbolPrices[t.Symbol] = t.Price
	}

	return t
}

// DeleteTrade removes a trade by ID
func DeleteTrade(id string) bool {
	mu.Lock()
	defer mu.Unlock()
	for i, t := range trades {
		if t.ID == id {
			trades = append(trades[:i], trades[i+1:]...)
			return true
		}
	}
	return false
}

// SetSymbolPrice updates the current market price for a symbol and records the update time.
func SetSymbolPrice(symbol string, price float64) {
	mu.Lock()
	defer mu.Unlock()
	symbolPrices[symbol] = price
	symbolLastUpdated[symbol] = time.Now()
}

// computeHoldings aggregates all trades into per-symbol holdings.
// Must be called with mu held (at least RLock).
func computeHoldings() []Holding {
	type accumulator struct {
		name         string
		buyQty       float64
		buyTotalCost float64 // sum of (qty*price + fee) for all buys
		sellQty      float64
	}
	// Preserve insertion order for stable output
	order := []string{}
	acc := map[string]*accumulator{}

	for _, t := range trades {
		if _, exists := acc[t.Symbol]; !exists {
			acc[t.Symbol] = &accumulator{name: t.Name}
			order = append(order, t.Symbol)
		}
		a := acc[t.Symbol]
		switch t.Type {
		case "buy":
			a.buyQty += t.Quantity
			a.buyTotalCost += t.Quantity*t.Price + t.Fee
		case "sell":
			a.sellQty += t.Quantity
		}
	}

	var holdings []Holding
	for _, sym := range order {
		a := acc[sym]
		netQty := a.buyQty - a.sellQty
		if netQty <= 0 {
			continue
		}

		var avgCost float64
		if a.buyQty > 0 {
			avgCost = a.buyTotalCost / a.buyQty
		}

		currentPrice := symbolPrices[sym]
		value := netQty * currentPrice
		totalCost := netQty * avgCost
		gainLoss := value - totalCost
		var gainLossPct float64
		if totalCost > 0 {
			gainLossPct = gainLoss / totalCost * 100
		}

		lastUpdated := ""
		if t, ok := symbolLastUpdated[sym]; ok {
			lastUpdated = t.UTC().Format(time.RFC3339)
		}

		holdings = append(holdings, Holding{
			Symbol:       sym,
			Name:         a.name,
			Quantity:     netQty,
			AvgCost:      avgCost,
			CurrentPrice: currentPrice,
			Value:        value,
			TotalCost:    totalCost,
			GainLoss:     gainLoss,
			GainLossPct:  gainLossPct,
			LastUpdated:  lastUpdated,
		})
	}
	return holdings
}

// GetHoldings returns aggregated per-symbol holdings derived from all trades
func GetHoldings() []Holding {
	mu.RLock()
	defer mu.RUnlock()
	return computeHoldings()
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

	// Prefer trade-based holdings for portfolio value; fall back to flat assets.
	var portfolioValue float64
	if len(trades) > 0 {
		for _, h := range computeHoldings() {
			portfolioValue += h.Value
		}
	} else {
		for _, a := range assets {
			portfolioValue += a.Value
		}
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
	if len(trades) > 0 {
		for _, h := range computeHoldings() {
			portfolioValue += h.Value
		}
	} else {
		for _, a := range assets {
			portfolioValue += a.Value
		}
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

// GetPriceConfig returns the current poller configuration
func GetPriceConfig() PriceConfig {
	mu.RLock()
	defer mu.RUnlock()
	return priceConfig
}

// SetPriceConfig replaces the poller configuration
func SetPriceConfig(cfg PriceConfig) {
	mu.Lock()
	defer mu.Unlock()
	if cfg.IntervalSeconds < 60 {
		cfg.IntervalSeconds = 60
	}
	priceConfig = cfg
}

// GetTrackedSymbols returns the list of symbols that have prices tracked
func GetTrackedSymbols() []string {
	mu.RLock()
	defer mu.RUnlock()
	syms := make([]string, 0, len(symbolPrices))
	for s := range symbolPrices {
		syms = append(syms, s)
	}
	return syms
}
