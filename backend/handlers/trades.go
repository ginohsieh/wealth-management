package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetTrades returns all portfolio trades, newest date first.
func GetTrades(c *gin.Context) {
	trades, err := store.GetTrades()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	// Sort descending by date (insertion sort on small slice)
	for i := 1; i < len(trades); i++ {
		for j := i; j > 0 && trades[j].Date > trades[j-1].Date; j-- {
			trades[j], trades[j-1] = trades[j-1], trades[j]
		}
	}
	c.JSON(http.StatusOK, trades)
}

// CreateTrade records a new buy or sell portfolio trade.
// Fee and Tax are auto-calculated from market rules when left as 0.
func CreateTrade(c *gin.Context) {
	var trade store.PortfolioTrade
	if err := c.ShouldBindJSON(&trade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if trade.Type != "buy" && trade.Type != "sell" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'buy' or 'sell'"})
		return
	}
	if trade.Quantity <= 0 || trade.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity and price must be positive"})
		return
	}
	created, err := store.CreateTrade(trade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// DeleteTrade removes a portfolio trade by ID.
func DeleteTrade(c *gin.Context) {
	ok, err := store.DeleteTrade(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "trade not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "trade deleted"})
}

// GetHoldings returns aggregated per-symbol positions derived from all trades.
func GetHoldings(c *gin.Context) {
	holdings, err := store.GetHoldings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if holdings == nil {
		holdings = []store.Holding{}
	}
	c.JSON(http.StatusOK, holdings)
}

// UpdateSymbolPrice sets the current market price for a symbol.
// Body: { "price": 123.45 }
func UpdateSymbolPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	var body struct {
		Price float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be a positive number"})
		return
	}
	if err := store.SetSymbolPrice(symbol, body.Price); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"symbol": symbol, "price": body.Price})
}

// CalcFeeAndTax returns the computed fee and tax for a potential trade without saving it.
// Query params: market, type (buy/sell), value (trade value = qty * price)
func CalcFeeAndTax(c *gin.Context) {
	var body struct {
		Market     string  `json:"market"`
		Type       string  `json:"type"`
		TradeValue float64 `json:"trade_value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fee, tax := store.CalcFeeAndTax(body.Market, body.Type, body.TradeValue)
	c.JSON(http.StatusOK, gin.H{"fee": fee, "tax": tax})
}
