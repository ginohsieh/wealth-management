package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetAssets returns all assets, optionally filtered by ?account_id=
func GetAssets(c *gin.Context) {
	assets, err := store.GetAssets(c.Query("account_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if assets == nil {
		assets = []store.Asset{}
	}
	c.JSON(http.StatusOK, assets)
}

// GetAsset returns a single asset by ID
func GetAsset(c *gin.Context) {
	asset, ok, err := store.GetAssetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	c.JSON(http.StatusOK, asset)
}

// CreateAccountAsset creates a new asset under an account
func CreateAccountAsset(c *gin.Context) {
	var asset store.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := store.CreateAsset(asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// UpdateAccountAsset replaces an existing asset
func UpdateAccountAsset(c *gin.Context) {
	var asset store.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, ok, err := store.UpdateAsset(c.Param("id"), asset)
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

// DeleteAccountAsset removes an asset by ID
func DeleteAccountAsset(c *gin.Context) {
	ok, err := store.DeleteAsset(c.Param("id"))
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

// GetAssetDailyValues returns daily value history for an asset
func GetAssetDailyValues(c *gin.Context) {
	values, err := store.GetAssetDailyValues(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if values == nil {
		values = []store.AssetDailyValue{}
	}
	c.JSON(http.StatusOK, values)
}

// UpsertAssetDailyValue records a daily value for an asset
func UpsertAssetDailyValue(c *gin.Context) {
	var v store.AssetDailyValue
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v.AssetID = c.Param("id")
	saved, err := store.UpsertAssetDailyValue(v)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusOK, saved)
}
