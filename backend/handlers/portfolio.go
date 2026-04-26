package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetPortfolio returns all legacy portfolio positions
func GetPortfolio(c *gin.Context) {
	assets, err := store.GetPortfolioPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// CreateAsset adds a new legacy portfolio position
func CreateAsset(c *gin.Context) {
	var asset store.PortfolioPosition
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := store.CreatePortfolioPosition(asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// UpdateAsset replaces an existing legacy portfolio position
func UpdateAsset(c *gin.Context) {
	var asset store.PortfolioPosition
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, ok, err := store.UpdatePortfolioPosition(c.Param("id"), asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteAsset removes a legacy portfolio position
func DeleteAsset(c *gin.Context) {
	ok, err := store.DeletePortfolioPosition(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "asset deleted"})
}
