package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetTransactions returns all transactions, with optional account_id filter
func GetTransactions(c *gin.Context) {
	txns, err := store.GetTransactions(c.Query("account_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if txns == nil {
		txns = []store.Transaction{}
	}
	c.JSON(http.StatusOK, txns)
}

// CreateTransaction adds a new transaction.
// For investment subtypes (stock_buy / stock_sell) the company name is resolved
// automatically from the symbol_names cache or the Yahoo Finance API.
func CreateTransaction(c *gin.Context) {
	var t store.Transaction
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve company name for investment transactions
	if (t.Subtype == "stock_buy" || t.Subtype == "stock_sell") && t.Symbol != "" {
		t.Symbol = strings.ToUpper(t.Symbol)
		// Seed a description if none provided
		if t.Description == "" {
			action := "Buy"
			if t.Subtype == "stock_sell" {
				action = "Sell"
			}
			t.Description = action + " " + t.Symbol
		}
	}

	created, err := store.CreateTransaction(t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}
