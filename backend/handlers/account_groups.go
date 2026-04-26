package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

func GetAccountGroups(c *gin.Context) {
	groups, err := store.GetAccountGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if groups == nil {
		groups = []store.AccountGroup{}
	}
	c.JSON(http.StatusOK, groups)
}

func GetAccountGroup(c *gin.Context) {
	group, ok, err := store.GetAccountGroupByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func CreateAccountGroup(c *gin.Context) {
	var g store.AccountGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if g.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	created, err := store.CreateAccountGroup(g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func UpdateAccountGroup(c *gin.Context) {
	var g store.AccountGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, ok, err := store.UpdateAccountGroup(c.Param("id"), g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteAccountGroup(c *gin.Context) {
	ok, err := store.DeleteAccountGroup(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

func SetGroupMembers(c *gin.Context) {
	var body struct {
		AccountIDs []string `json:"account_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.AccountIDs == nil {
		body.AccountIDs = []string{}
	}
	if err := store.SetGroupMembers(c.Param("id"), body.AccountIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	group, _, _ := store.GetAccountGroupByID(c.Param("id"))
	c.JSON(http.StatusOK, group)
}

func GetGroupStats(c *gin.Context) {
	stats, ok, err := store.GetGroupStats(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
