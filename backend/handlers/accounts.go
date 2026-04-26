package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetAccounts returns all accounts
func GetAccounts(c *gin.Context) {
	c.JSON(http.StatusOK, store.GetAccounts())
}

// GetAccount returns a single account by ID
func GetAccount(c *gin.Context) {
	account, ok := store.GetAccountByID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, account)
}

// CreateAccount creates a new account
func CreateAccount(c *gin.Context) {
	var account store.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, store.CreateAccount(account))
}

// UpdateAccount replaces an existing account
func UpdateAccount(c *gin.Context) {
	var account store.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, ok := store.UpdateAccount(c.Param("id"), account)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteAccount removes an account by ID
func DeleteAccount(c *gin.Context) {
	if !store.DeleteAccount(c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
