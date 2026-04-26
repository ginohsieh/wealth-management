package handlers

import (
	"net/http"

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

// CreateTransaction adds a new transaction
func CreateTransaction(c *gin.Context) {
	var t store.Transaction
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := store.CreateTransaction(t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}
