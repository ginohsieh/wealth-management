package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetSummary returns the aggregate financial summary
func GetSummary(c *gin.Context) {
	c.JSON(http.StatusOK, store.GetSummary())
}
