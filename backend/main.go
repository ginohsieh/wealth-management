package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/handlers"
)

func main() {
	r := gin.Default()

	// Allow all origins for local development
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
	}))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.GET("/summary", handlers.GetSummary)

		api.GET("/accounts", handlers.GetAccounts)
		api.POST("/accounts", handlers.CreateAccount)
		api.GET("/accounts/:id", handlers.GetAccount)
		api.PUT("/accounts/:id", handlers.UpdateAccount)
		api.DELETE("/accounts/:id", handlers.DeleteAccount)

		api.GET("/transactions", handlers.GetTransactions)
		api.POST("/transactions", handlers.CreateTransaction)

		api.GET("/portfolio", handlers.GetPortfolio)
		api.POST("/portfolio", handlers.CreateAsset)
		api.PUT("/portfolio/:id", handlers.UpdateAsset)
		api.DELETE("/portfolio/:id", handlers.DeleteAsset)

		api.GET("/snapshots", handlers.GetSnapshots)
		api.POST("/snapshots", handlers.RecordSnapshot)
	}

	r.Run(":8080")
}
