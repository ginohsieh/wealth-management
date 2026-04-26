package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetPortfolio returns all portfolio assets
func GetPortfolio(c *gin.Context) {
	c.JSON(http.StatusOK, store.GetAssets())
}

// CreateAsset adds a new asset to the portfolio
func CreateAsset(c *gin.Context) {
	var asset store.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, store.CreateAsset(asset))
}

// UpdateAsset replaces an existing portfolio asset
func UpdateAsset(c *gin.Context) {
	var asset store.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, ok := store.UpdateAsset(c.Param("id"), asset)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteAsset removes an asset from the portfolio
func DeleteAsset(c *gin.Context) {
	if !store.DeleteAsset(c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "asset deleted"})
}
