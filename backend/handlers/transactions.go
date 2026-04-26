package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetTransactions returns all transactions, with optional account_id filter
func GetTransactions(c *gin.Context) {
	txns := store.GetTransactions(c.Query("account_id"))
	if txns == nil {
		txns = []store.Transaction{}
	}
	c.JSON(http.StatusOK, txns)
}

// CreateTransaction adds a new transaction
func CreateTransaction(c *gin.Context) {
	var t store.Transaction
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, store.CreateTransaction(t))
}
