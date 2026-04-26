package store

import (
"database/sql"
"fmt"
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
ID        string  `json:"id"`
Symbol    string  `json:"symbol"`
Name      string  `json:"name"`
Type      string  `json:"type"`   // "buy" or "sell"
Date      string  `json:"date"`
Quantity  float64 `json:"quantity"`
Price     float64 `json:"price"`
Market    string  `json:"market"` // "US", "TW", "HK", "OTHER"
Fee       float64 `json:"fee"`
Tax       float64 `json:"tax"`
Notes     string  `json:"notes"`
AccountID string  `json:"account_id"`
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
Source string `json:"source"`
// IntervalSeconds is how often to poll (minimum 60, default 300)
IntervalSeconds int `json:"interval_seconds"`
// APIKey is required for Alpha Vantage; ignored for Yahoo
APIKey string `json:"api_key"`
// Enabled turns polling on or off without changing other settings
Enabled bool `json:"enabled"`
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
ID               string  `json:"id"`
Date             string  `json:"date"`
PortfolioValue   float64 `json:"portfolio_value"`
TotalAssets      float64 `json:"total_assets"`
TotalLiabilities float64 `json:"total_liabilities"`
NetWorth         float64 `json:"net_worth"`
}

// AccountGroup is a named collection of accounts for aggregate reporting.
type AccountGroup struct {
ID          string   `json:"id"`
Name        string   `json:"name"`
Description string   `json:"description"`
AccountIDs  []string `json:"account_ids"`
}

// GroupStats aggregates balance and portfolio metrics for an account group.
type GroupStats struct {
GroupID        string  `json:"group_id"`
GroupName      string  `json:"group_name"`
TotalBalance   float64 `json:"total_balance"`
PortfolioValue float64 `json:"portfolio_value"`
TotalCost      float64 `json:"total_cost"`
UnrealizedGain float64 `json:"unrealized_gain"`
ROR            float64 `json:"ror"`
}

// db is the package-level database connection set by Init.
var db *sql.DB

// Init sets the database connection for the store. Must be called before any
// store function is used.
func Init(database *sql.DB) {
db = database
}

// ── Accounts ─────────────────────────────────────────────────────────────────

// GetAccounts returns all accounts ordered by ID.
func GetAccounts() ([]Account, error) {
rows, err := db.Query(`SELECT id, name, type, balance, currency FROM accounts ORDER BY id`)
if err != nil {
return nil, err
}
defer rows.Close()

result := []Account{}
for rows.Next() {
var a Account
var id int
if err := rows.Scan(&id, &a.Name, &a.Type, &a.Balance, &a.Currency); err != nil {
return nil, err
}
a.ID = fmt.Sprintf("%d", id)
result = append(result, a)
}
return result, rows.Err()
}

// GetAccountByID returns an account by its ID. The second return value is
// false when the account does not exist.
func GetAccountByID(id string) (Account, bool, error) {
var a Account
var dbID int
err := db.QueryRow(
`SELECT id, name, type, balance, currency FROM accounts WHERE id = $1`, id,
).Scan(&dbID, &a.Name, &a.Type, &a.Balance, &a.Currency)
if err == sql.ErrNoRows {
return Account{}, false, nil
}
if err != nil {
return Account{}, false, err
}
a.ID = fmt.Sprintf("%d", dbID)
return a, true, nil
}

// CreateAccount inserts a new account and returns it with a generated ID.
func CreateAccount(a Account) (Account, error) {
if a.Currency == "" {
a.Currency = "USD"
}
var id int
err := db.QueryRow(
`INSERT INTO accounts(name, type, balance, currency) VALUES($1,$2,$3,$4) RETURNING id`,
a.Name, a.Type, a.Balance, a.Currency,
).Scan(&id)
if err != nil {
return Account{}, err
}
a.ID = fmt.Sprintf("%d", id)
return a, nil
}

// UpdateAccount replaces an existing account.
func UpdateAccount(id string, a Account) (Account, bool, error) {
var dbID int
err := db.QueryRow(
`UPDATE accounts SET name=$2, type=$3, balance=$4, currency=$5 WHERE id=$1 RETURNING id`,
id, a.Name, a.Type, a.Balance, a.Currency,
).Scan(&dbID)
if err == sql.ErrNoRows {
return Account{}, false, nil
}
if err != nil {
return Account{}, false, err
}
a.ID = fmt.Sprintf("%d", dbID)
return a, true, nil
}

// DeleteAccount removes an account by ID.
func DeleteAccount(id string) (bool, error) {
res, err := db.Exec(`DELETE FROM accounts WHERE id=$1`, id)
if err != nil {
return false, err
}
n, _ := res.RowsAffected()
return n > 0, nil
}

// ── Transactions ──────────────────────────────────────────────────────────────

// GetTransactions returns all transactions, optionally filtered by account ID.
func GetTransactions(accountID string) ([]Transaction, error) {
var rows *sql.Rows
var err error
if accountID == "" {
rows, err = db.Query(
`SELECT id, account_id, date, description, amount, category, type
 FROM transactions ORDER BY id`)
} else {
rows, err = db.Query(
`SELECT id, account_id, date, description, amount, category, type
 FROM transactions WHERE account_id=$1 ORDER BY id`, accountID)
}
if err != nil {
return nil, err
}
defer rows.Close()

result := []Transaction{}
for rows.Next() {
var t Transaction
var id, accountIDInt int
if err := rows.Scan(&id, &accountIDInt, &t.Date, &t.Description,
&t.Amount, &t.Category, &t.Type); err != nil {
return nil, err
}
t.ID = fmt.Sprintf("%d", id)
t.AccountID = fmt.Sprintf("%d", accountIDInt)
result = append(result, t)
}
return result, rows.Err()
}

// CreateTransaction inserts a new transaction and returns it with a generated ID.
func CreateTransaction(t Transaction) (Transaction, error) {
var id int
err := db.QueryRow(
`INSERT INTO transactions(account_id, date, description, amount, category, type)
 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
t.AccountID, t.Date, t.Description, t.Amount, t.Category, t.Type,
).Scan(&id)
if err != nil {
return Transaction{}, err
}
t.ID = fmt.Sprintf("%d", id)
return t, nil
}

// ── Assets (legacy flat portfolio) ───────────────────────────────────────────

// GetAssets returns all portfolio assets.
func GetAssets() ([]Asset, error) {
rows, err := db.Query(
`SELECT id, symbol, name, quantity, purchase_price, current_price
 FROM assets ORDER BY id`)
if err != nil {
return nil, err
}
defer rows.Close()

result := []Asset{}
for rows.Next() {
var a Asset
var id int
if err := rows.Scan(&id, &a.Symbol, &a.Name, &a.Quantity,
&a.PurchasePrice, &a.CurrentPrice); err != nil {
return nil, err
}
a.ID = fmt.Sprintf("%d", id)
a = calculateAssetFields(a)
result = append(result, a)
}
return result, rows.Err()
}

// CreateAsset inserts a new asset and returns it with computed fields.
func CreateAsset(a Asset) (Asset, error) {
a = calculateAssetFields(a)
var id int
err := db.QueryRow(
`INSERT INTO assets(symbol, name, quantity, purchase_price, current_price)
 VALUES($1,$2,$3,$4,$5) RETURNING id`,
a.Symbol, a.Name, a.Quantity, a.PurchasePrice, a.CurrentPrice,
).Scan(&id)
if err != nil {
return Asset{}, err
}
a.ID = fmt.Sprintf("%d", id)
return a, nil
}

// UpdateAsset replaces an existing asset.
func UpdateAsset(id string, a Asset) (Asset, bool, error) {
a = calculateAssetFields(a)
var dbID int
err := db.QueryRow(
`UPDATE assets SET symbol=$2, name=$3, quantity=$4, purchase_price=$5, current_price=$6
 WHERE id=$1 RETURNING id`,
id, a.Symbol, a.Name, a.Quantity, a.PurchasePrice, a.CurrentPrice,
).Scan(&dbID)
if err == sql.ErrNoRows {
return Asset{}, false, nil
}
if err != nil {
return Asset{}, false, err
}
a.ID = fmt.Sprintf("%d", dbID)
return a, true, nil
}

// DeleteAsset removes an asset by ID.
func DeleteAsset(id string) (bool, error) {
res, err := db.Exec(`DELETE FROM assets WHERE id=$1`, id)
if err != nil {
return false, err
}
n, _ := res.RowsAffected()
return n > 0, nil
}

func calculateAssetFields(a Asset) Asset {
a.Value = a.Quantity * a.CurrentPrice
a.GainLoss = a.Value - (a.Quantity * a.PurchasePrice)
if a.PurchasePrice > 0 {
a.GainLossPct = (a.CurrentPrice - a.PurchasePrice) / a.PurchasePrice * 100
}
return a
}

// ── CalcFeeAndTax (pure, no DB) ───────────────────────────────────────────────

// CalcFeeAndTax returns the broker fee and transaction tax for a trade based on
// market rules. Values are in the same currency as the trade.
//
// Rules applied:
//
//US    – fee=0 (commission-free brokers), tax=0
//TW    – fee=0.1425% of trade value (broker commission, both sides),
//         tax=0.3% of sell value (securities transaction tax, sell only)
//HK    – fee=0.25% of trade value (broker commission, both sides),
//         stamp duty=0.1% of trade value (both sides)
//OTHER – fee=0, tax=0
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

// ── Portfolio Trades ──────────────────────────────────────────────────────────

// GetTrades returns all portfolio trades, optionally filtered by account ID.
func GetTrades(accountID string) ([]PortfolioTrade, error) {
var rows *sql.Rows
var err error
if accountID == "" {
rows, err = db.Query(
`SELECT id, symbol, name, type, date, quantity, price, market, fee, tax, notes, account_id
 FROM portfolio_trades ORDER BY id`)
} else {
rows, err = db.Query(
`SELECT id, symbol, name, type, date, quantity, price, market, fee, tax, notes, account_id
 FROM portfolio_trades WHERE account_id = $1 ORDER BY id`, accountID)
}
if err != nil {
return nil, err
}
defer rows.Close()

result := []PortfolioTrade{}
for rows.Next() {
var t PortfolioTrade
var id int
var accountIDNull sql.NullInt64
if err := rows.Scan(&id, &t.Symbol, &t.Name, &t.Type, &t.Date,
&t.Quantity, &t.Price, &t.Market, &t.Fee, &t.Tax, &t.Notes, &accountIDNull); err != nil {
return nil, err
}
t.ID = fmt.Sprintf("%d", id)
if accountIDNull.Valid {
t.AccountID = fmt.Sprintf("%d", accountIDNull.Int64)
}
result = append(result, t)
}
return result, rows.Err()
}

// CreateTrade records a new buy or sell trade. Fee and Tax are auto-calculated
// when the caller leaves them as 0.
func CreateTrade(t PortfolioTrade) (PortfolioTrade, error) {
tradeValue := t.Quantity * t.Price
autoFee, autoTax := CalcFeeAndTax(t.Market, t.Type, tradeValue)
if t.Fee == 0 {
t.Fee = autoFee
}
if t.Tax == 0 {
t.Tax = autoTax
}

var accountID interface{}
if t.AccountID != "" {
accountID = t.AccountID
} else {
accountID = nil
}

var id int
err := db.QueryRow(
`INSERT INTO portfolio_trades(symbol, name, type, date, quantity, price, market, fee, tax, notes, account_id)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`,
t.Symbol, t.Name, t.Type, t.Date, t.Quantity, t.Price,
t.Market, t.Fee, t.Tax, t.Notes, accountID,
).Scan(&id)
if err != nil {
return PortfolioTrade{}, err
}
t.ID = fmt.Sprintf("%d", id)

// Initialize price for new symbols (do not overwrite an existing price).
if _, err := db.Exec(
`INSERT INTO symbol_prices(symbol, price) VALUES($1,$2) ON CONFLICT DO NOTHING`,
t.Symbol, t.Price,
); err != nil {
return t, err
}
return t, nil
}

// DeleteTrade removes a trade by ID.
func DeleteTrade(id string) (bool, error) {
res, err := db.Exec(`DELETE FROM portfolio_trades WHERE id=$1`, id)
if err != nil {
return false, err
}
n, _ := res.RowsAffected()
return n > 0, nil
}

// SetSymbolPrice updates the current market price for a symbol.
func SetSymbolPrice(symbol string, price float64) error {
_, err := db.Exec(
`INSERT INTO symbol_prices(symbol, price, last_updated)
 VALUES($1,$2,NOW())
 ON CONFLICT(symbol) DO UPDATE SET price=$2, last_updated=NOW()`,
symbol, price,
)
return err
}

// GetTrackedSymbols returns the list of symbols that have prices tracked.
func GetTrackedSymbols() ([]string, error) {
rows, err := db.Query(`SELECT symbol FROM symbol_prices`)
if err != nil {
return nil, err
}
defer rows.Close()

var syms []string
for rows.Next() {
var s string
if err := rows.Scan(&s); err != nil {
return nil, err
}
syms = append(syms, s)
}
return syms, rows.Err()
}

// GetSymbolName looks up a company name from the symbol_names cache.
// Returns (name, true, nil) when found, ("", false, nil) when not cached.
func GetSymbolName(symbol string) (string, bool, error) {
var name string
err := db.QueryRow(`SELECT name FROM symbol_names WHERE symbol = $1`, symbol).Scan(&name)
if err == sql.ErrNoRows {
return "", false, nil
}
if err != nil {
return "", false, err
}
return name, true, nil
}

// SetSymbolName upserts a symbol → company name mapping in the symbol_names cache.
func SetSymbolName(symbol, name string) error {
_, err := db.Exec(
`INSERT INTO symbol_names(symbol, name, fetched_at)
 VALUES($1,$2,NOW())
 ON CONFLICT(symbol) DO UPDATE SET name=$2, fetched_at=NOW()`,
symbol, name,
)
return err
}

// computeHoldings loads trades and symbol prices from the DB and aggregates
// them into per-symbol holdings. Optionally filtered by accountID.
func computeHoldings(accountID string) ([]Holding, error) {
var rows *sql.Rows
var err error
if accountID == "" {
rows, err = db.Query(
`SELECT symbol, name, type, quantity, price, fee FROM portfolio_trades ORDER BY id`)
} else {
rows, err = db.Query(
`SELECT symbol, name, type, quantity, price, fee FROM portfolio_trades WHERE account_id = $1 ORDER BY id`, accountID)
}
if err != nil {
return nil, err
}
defer rows.Close()

type accumulator struct {
name         string
buyQty       float64
buyTotalCost float64
sellQty      float64
}
order := []string{}
acc := map[string]*accumulator{}

for rows.Next() {
var sym, name, tradeType string
var qty, price, fee float64
if err := rows.Scan(&sym, &name, &tradeType, &qty, &price, &fee); err != nil {
return nil, err
}
if _, exists := acc[sym]; !exists {
acc[sym] = &accumulator{name: name}
order = append(order, sym)
}
a := acc[sym]
switch tradeType {
case "buy":
a.buyQty += qty
a.buyTotalCost += qty*price + fee
case "sell":
a.sellQty += qty
}
}
if err := rows.Err(); err != nil {
return nil, err
}

// Load symbol prices and last-updated timestamps
type priceRow struct {
price       float64
lastUpdated *time.Time
}
prices := map[string]priceRow{}
pRows, err := db.Query(`SELECT symbol, price, last_updated FROM symbol_prices`)
if err != nil {
return nil, err
}
defer pRows.Close()
for pRows.Next() {
var sym string
var p float64
var lu *time.Time
if err := pRows.Scan(&sym, &p, &lu); err != nil {
return nil, err
}
prices[sym] = priceRow{price: p, lastUpdated: lu}
}
if err := pRows.Err(); err != nil {
return nil, err
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
pr := prices[sym]
value := netQty * pr.price
totalCost := netQty * avgCost
gainLoss := value - totalCost
var gainLossPct float64
if totalCost > 0 {
gainLossPct = gainLoss / totalCost * 100
}
lastUpdated := ""
if pr.lastUpdated != nil {
lastUpdated = pr.lastUpdated.UTC().Format(time.RFC3339)
}
holdings = append(holdings, Holding{
Symbol:       sym,
Name:         a.name,
Quantity:     netQty,
AvgCost:      avgCost,
CurrentPrice: pr.price,
Value:        value,
TotalCost:    totalCost,
GainLoss:     gainLoss,
GainLossPct:  gainLossPct,
LastUpdated:  lastUpdated,
})
}
return holdings, nil
}

// GetHoldings returns aggregated per-symbol holdings, optionally filtered by account.
func GetHoldings(accountID string) ([]Holding, error) {
return computeHoldings(accountID)
}

// ── Summary ───────────────────────────────────────────────────────────────────

// GetSummary returns aggregate financial metrics.
func GetSummary() (Summary, error) {
// Account totals
aRows, err := db.Query(`SELECT balance FROM accounts`)
if err != nil {
return Summary{}, err
}
defer aRows.Close()
var totalAssets, totalLiabilities float64
for aRows.Next() {
var bal float64
if err := aRows.Scan(&bal); err != nil {
return Summary{}, err
}
if bal >= 0 {
totalAssets += bal
} else {
totalLiabilities += -bal
}
}
if err := aRows.Err(); err != nil {
return Summary{}, err
}

// Portfolio value — prefer trade-based holdings
var portfolioValue float64
var tradeCount int
if err := db.QueryRow(`SELECT COUNT(*) FROM portfolio_trades`).Scan(&tradeCount); err != nil {
return Summary{}, err
}
if tradeCount > 0 {
holdings, err := computeHoldings("")
if err != nil {
return Summary{}, err
}
for _, h := range holdings {
portfolioValue += h.Value
}
} else {
if err := db.QueryRow(
`SELECT COALESCE(SUM(quantity * current_price), 0) FROM assets`,
).Scan(&portfolioValue); err != nil {
return Summary{}, err
}
}

// Monthly income / expense from transactions
var monthlyIncome, monthlyExpenses float64
tRows, err := db.Query(`SELECT type, amount FROM transactions`)
if err != nil {
return Summary{}, err
}
defer tRows.Close()
for tRows.Next() {
var typ string
var amount float64
if err := tRows.Scan(&typ, &amount); err != nil {
return Summary{}, err
}
switch typ {
case "income":
monthlyIncome += amount
case "expense":
monthlyExpenses += -amount
}
}
if err := tRows.Err(); err != nil {
return Summary{}, err
}

return Summary{
TotalAssets:      totalAssets,
TotalLiabilities: totalLiabilities,
NetWorth:         totalAssets - totalLiabilities,
MonthlyIncome:    monthlyIncome,
MonthlyExpenses:  monthlyExpenses,
MonthlySavings:   monthlyIncome - monthlyExpenses,
PortfolioValue:   portfolioValue,
}, nil
}

// ── Net Worth Snapshots ───────────────────────────────────────────────────────

// GetSnapshots returns all recorded net worth snapshots.
func GetSnapshots() ([]NetWorthSnapshot, error) {
rows, err := db.Query(
`SELECT id, date, portfolio_value, total_assets, total_liabilities, net_worth
 FROM net_worth_snapshots ORDER BY date`)
if err != nil {
return nil, err
}
defer rows.Close()

result := []NetWorthSnapshot{}
for rows.Next() {
var s NetWorthSnapshot
var id int
if err := rows.Scan(&id, &s.Date, &s.PortfolioValue,
&s.TotalAssets, &s.TotalLiabilities, &s.NetWorth); err != nil {
return nil, err
}
s.ID = fmt.Sprintf("%d", id)
result = append(result, s)
}
return result, rows.Err()
}

// RecordSnapshot captures the current financial state for the given date.
// If a snapshot already exists for that date it is replaced.
func RecordSnapshot(date string) (NetWorthSnapshot, error) {
var totalAssets, totalLiabilities float64
aRows, err := db.Query(`SELECT balance FROM accounts`)
if err != nil {
return NetWorthSnapshot{}, err
}
defer aRows.Close()
for aRows.Next() {
var bal float64
if err := aRows.Scan(&bal); err != nil {
return NetWorthSnapshot{}, err
}
if bal >= 0 {
totalAssets += bal
} else {
totalLiabilities += -bal
}
}
if err := aRows.Err(); err != nil {
return NetWorthSnapshot{}, err
}

var portfolioValue float64
var tradeCount int
if err := db.QueryRow(`SELECT COUNT(*) FROM portfolio_trades`).Scan(&tradeCount); err != nil {
return NetWorthSnapshot{}, err
}
if tradeCount > 0 {
holdings, err := computeHoldings("")
if err != nil {
return NetWorthSnapshot{}, err
}
for _, h := range holdings {
portfolioValue += h.Value
}
} else {
if err := db.QueryRow(
`SELECT COALESCE(SUM(quantity * current_price), 0) FROM assets`,
).Scan(&portfolioValue); err != nil {
return NetWorthSnapshot{}, err
}
}

var snap NetWorthSnapshot
var id int
err = db.QueryRow(`
INSERT INTO net_worth_snapshots(date, portfolio_value, total_assets, total_liabilities, net_worth)
VALUES($1,$2,$3,$4,$5)
ON CONFLICT(date) DO UPDATE
  SET portfolio_value=$2, total_assets=$3, total_liabilities=$4, net_worth=$5
RETURNING id`,
date, portfolioValue, totalAssets, totalLiabilities, totalAssets-totalLiabilities,
).Scan(&id)
if err != nil {
return NetWorthSnapshot{}, err
}
snap.ID = fmt.Sprintf("%d", id)
snap.Date = date
snap.PortfolioValue = portfolioValue
snap.TotalAssets = totalAssets
snap.TotalLiabilities = totalLiabilities
snap.NetWorth = totalAssets - totalLiabilities
return snap, nil
}

// ── Price Config ──────────────────────────────────────────────────────────────

// GetPriceConfig returns the current poller configuration from the DB.
func GetPriceConfig() (PriceConfig, error) {
var cfg PriceConfig
err := db.QueryRow(
`SELECT source, interval_seconds, api_key, enabled FROM price_config WHERE id=1`,
).Scan(&cfg.Source, &cfg.IntervalSeconds, &cfg.APIKey, &cfg.Enabled)
if err == sql.ErrNoRows {
// Return sensible defaults if the row doesn't exist yet
return PriceConfig{Source: "yahoo", IntervalSeconds: 300}, nil
}
return cfg, err
}

// SetPriceConfig upserts the poller configuration into the DB.
func SetPriceConfig(cfg PriceConfig) error {
if cfg.IntervalSeconds < 60 {
cfg.IntervalSeconds = 60
}
_, err := db.Exec(`
INSERT INTO price_config(id, source, interval_seconds, api_key, enabled)
VALUES(1,$1,$2,$3,$4)
ON CONFLICT(id) DO UPDATE
  SET source=$1, interval_seconds=$2, api_key=$3, enabled=$4`,
cfg.Source, cfg.IntervalSeconds, cfg.APIKey, cfg.Enabled,
)
return err
}

// ── Account Groups ────────────────────────────────────────────────────────────

func GetAccountGroups() ([]AccountGroup, error) {
rows, err := db.Query(`SELECT id, name, description FROM account_groups ORDER BY id`)
if err != nil {
return nil, err
}
defer rows.Close()

var groups []AccountGroup
groupIndex := map[string]int{}
for rows.Next() {
var g AccountGroup
var id int
if err := rows.Scan(&id, &g.Name, &g.Description); err != nil {
return nil, err
}
g.ID = fmt.Sprintf("%d", id)
g.AccountIDs = []string{}
groupIndex[g.ID] = len(groups)
groups = append(groups, g)
}
if err := rows.Err(); err != nil {
return nil, err
}

mRows, err := db.Query(`SELECT group_id, account_id FROM account_group_members ORDER BY group_id, account_id`)
if err != nil {
return nil, err
}
defer mRows.Close()
for mRows.Next() {
var gid, aid int
if err := mRows.Scan(&gid, &aid); err != nil {
return nil, err
}
gidStr := fmt.Sprintf("%d", gid)
if idx, ok := groupIndex[gidStr]; ok {
groups[idx].AccountIDs = append(groups[idx].AccountIDs, fmt.Sprintf("%d", aid))
}
}
return groups, mRows.Err()
}

func GetAccountGroupByID(id string) (AccountGroup, bool, error) {
var g AccountGroup
var dbID int
err := db.QueryRow(`SELECT id, name, description FROM account_groups WHERE id = $1`, id).
Scan(&dbID, &g.Name, &g.Description)
if err == sql.ErrNoRows {
return AccountGroup{}, false, nil
}
if err != nil {
return AccountGroup{}, false, err
}
g.ID = fmt.Sprintf("%d", dbID)
g.AccountIDs = []string{}

mRows, err := db.Query(`SELECT account_id FROM account_group_members WHERE group_id = $1 ORDER BY account_id`, id)
if err != nil {
return AccountGroup{}, false, err
}
defer mRows.Close()
for mRows.Next() {
var aid int
if err := mRows.Scan(&aid); err != nil {
return AccountGroup{}, false, err
}
g.AccountIDs = append(g.AccountIDs, fmt.Sprintf("%d", aid))
}
return g, true, mRows.Err()
}

func CreateAccountGroup(g AccountGroup) (AccountGroup, error) {
var id int
err := db.QueryRow(
`INSERT INTO account_groups(name, description) VALUES($1,$2) RETURNING id`,
g.Name, g.Description,
).Scan(&id)
if err != nil {
return AccountGroup{}, err
}
g.ID = fmt.Sprintf("%d", id)
if g.AccountIDs == nil {
g.AccountIDs = []string{}
}
return g, nil
}

func UpdateAccountGroup(id string, g AccountGroup) (AccountGroup, bool, error) {
var dbID int
err := db.QueryRow(
`UPDATE account_groups SET name=$2, description=$3 WHERE id=$1 RETURNING id`,
id, g.Name, g.Description,
).Scan(&dbID)
if err == sql.ErrNoRows {
return AccountGroup{}, false, nil
}
if err != nil {
return AccountGroup{}, false, err
}
g.ID = fmt.Sprintf("%d", dbID)
updated, _, err := GetAccountGroupByID(g.ID)
return updated, true, err
}

func DeleteAccountGroup(id string) (bool, error) {
res, err := db.Exec(`DELETE FROM account_groups WHERE id=$1`, id)
if err != nil {
return false, err
}
n, _ := res.RowsAffected()
return n > 0, nil
}

func SetGroupMembers(groupID string, accountIDs []string) error {
tx, err := db.Begin()
if err != nil {
return err
}
defer tx.Rollback()

if _, err := tx.Exec(`DELETE FROM account_group_members WHERE group_id = $1`, groupID); err != nil {
return err
}
for _, aid := range accountIDs {
if _, err := tx.Exec(
`INSERT INTO account_group_members(group_id, account_id) VALUES($1,$2)`,
groupID, aid,
); err != nil {
return err
}
}
return tx.Commit()
}

func GetGroupStats(groupID string) (GroupStats, bool, error) {
var g AccountGroup
var dbID int
err := db.QueryRow(`SELECT id, name FROM account_groups WHERE id = $1`, groupID).
Scan(&dbID, &g.Name)
if err == sql.ErrNoRows {
return GroupStats{}, false, nil
}
if err != nil {
return GroupStats{}, false, err
}
g.ID = fmt.Sprintf("%d", dbID)

mRows, err := db.Query(`SELECT account_id FROM account_group_members WHERE group_id = $1`, groupID)
if err != nil {
return GroupStats{}, false, err
}
defer mRows.Close()
var accountIDs []string
for mRows.Next() {
var aid int
if err := mRows.Scan(&aid); err != nil {
return GroupStats{}, false, err
}
accountIDs = append(accountIDs, fmt.Sprintf("%d", aid))
}
if err := mRows.Err(); err != nil {
return GroupStats{}, false, err
}

var totalBalance float64
for _, aid := range accountIDs {
var bal float64
if err := db.QueryRow(`SELECT COALESCE(balance, 0) FROM accounts WHERE id = $1`, aid).Scan(&bal); err != nil && err != sql.ErrNoRows {
return GroupStats{}, false, err
}
totalBalance += bal
}

type symHolding struct {
value     float64
totalCost float64
}
holdingsMap := map[string]symHolding{}
for _, aid := range accountIDs {
holdings, err := computeHoldings(aid)
if err != nil {
return GroupStats{}, false, err
}
for _, h := range holdings {
existing := holdingsMap[h.Symbol]
existing.value += h.Value
existing.totalCost += h.TotalCost
holdingsMap[h.Symbol] = existing
}
}

var portfolioValue, totalCost float64
for _, sh := range holdingsMap {
portfolioValue += sh.value
totalCost += sh.totalCost
}

unrealizedGain := portfolioValue - totalCost
var ror float64
if totalCost > 0 {
ror = unrealizedGain / totalCost * 100
}

return GroupStats{
GroupID:        g.ID,
GroupName:      g.Name,
TotalBalance:   totalBalance,
PortfolioValue: portfolioValue,
TotalCost:      totalCost,
UnrealizedGain: unrealizedGain,
ROR:            ror,
}, true, nil
}
