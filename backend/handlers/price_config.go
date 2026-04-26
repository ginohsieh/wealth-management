package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/poller"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetPriceConfig returns the current poller configuration.
func GetPriceConfig(c *gin.Context) {
	c.JSON(http.StatusOK, store.GetPriceConfig())
}

// UpdatePriceConfig replaces the poller configuration and restarts the poller.
// Body: { "source": "yahoo"|"alphavantage", "interval_seconds": 300, "api_key": "", "enabled": true }
func UpdatePriceConfig(c *gin.Context) {
	var cfg store.PriceConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	store.SetPriceConfig(cfg)
	poller.Restart()
	c.JSON(http.StatusOK, store.GetPriceConfig())
}

// ManualRefresh immediately fetches the latest prices for all tracked symbols.
func ManualRefresh(c *gin.Context) {
	updated := poller.Refresh()
	c.JSON(http.StatusOK, gin.H{
		"updated": updated,
		"count":   len(updated),
	})
}
