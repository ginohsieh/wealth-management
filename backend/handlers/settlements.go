package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetSettlements returns pending settlement records.
// Optional query param: ?status=pending|settled (default: all)
func GetSettlements(c *gin.Context) {
	status := c.Query("status")
	settlements, err := store.GetPendingSettlements(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if settlements == nil {
		settlements = []store.PendingSettlement{}
	}
	c.JSON(http.StatusOK, settlements)
}

// SettleOne forces immediate settlement of a single pending settlement record.
// Useful for testing or manual reconciliation.
func SettleOne(c *gin.Context) {
	id := c.Param("id")
	settled, found, err := store.ForceSettle(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "settlement not found or already settled"})
		return
	}
	c.JSON(http.StatusOK, settled)
}

// RunSettlements triggers the batch settlement processor immediately.
// Returns the count of settlements processed.
func RunSettlements(c *gin.Context) {
	n, err := store.SettlePendingTransactions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "settlement run failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"settled": n})
}
