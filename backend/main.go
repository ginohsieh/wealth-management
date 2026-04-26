package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/db"
	"github.com/ginohsieh/wealth-management/backend/handlers"
	"github.com/ginohsieh/wealth-management/backend/poller"
	"github.com/ginohsieh/wealth-management/backend/store"
)

func main() {
	// Initialize database connection
	database, err := db.Open()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create schema and seed initial data when the database is empty
	if err := db.Migrate(database); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	// Wire the database into the store layer
	store.Init(database)

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

	// Start background price poller (honours the config's Enabled flag)
	poller.Start()
	// Start background settlement runner (processes due settlements daily)
	poller.StartSettlementRunner()

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

		// Transaction-based portfolio
		api.GET("/portfolio/trades", handlers.GetTrades)
		api.POST("/portfolio/trades", handlers.CreateTrade)
		api.DELETE("/portfolio/trades/:id", handlers.DeleteTrade)
		api.GET("/portfolio/holdings", handlers.GetHoldings)
		api.PUT("/portfolio/holdings/:symbol/price", handlers.UpdateSymbolPrice)
		api.POST("/portfolio/calc-fees", handlers.CalcFeeAndTax)

		// Symbol name lookup (cache → Yahoo Finance)
		api.GET("/symbol/:symbol/name", handlers.LookupSymbolName)

		// Price polling configuration
		api.GET("/price-config", handlers.GetPriceConfig)
		api.PUT("/price-config", handlers.UpdatePriceConfig)
		api.POST("/price-refresh", handlers.ManualRefresh)

		api.GET("/snapshots", handlers.GetSnapshots)
		api.POST("/snapshots", handlers.RecordSnapshot)

		api.GET("/account-groups", handlers.GetAccountGroups)
		api.POST("/account-groups", handlers.CreateAccountGroup)
		api.GET("/account-groups/:id", handlers.GetAccountGroup)
		api.PUT("/account-groups/:id", handlers.UpdateAccountGroup)
		api.DELETE("/account-groups/:id", handlers.DeleteAccountGroup)
		api.PUT("/account-groups/:id/members", handlers.SetGroupMembers)
		api.GET("/account-groups/:id/stats", handlers.GetGroupStats)

		// Assets (account → transaction bridge)
		api.GET("/assets", handlers.GetAssets)
		api.POST("/assets", handlers.CreateAccountAsset)
		api.GET("/assets/:id", handlers.GetAsset)
		api.PUT("/assets/:id", handlers.UpdateAccountAsset)
		api.DELETE("/assets/:id", handlers.DeleteAccountAsset)
		api.GET("/assets/:id/daily-values", handlers.GetAssetDailyValues)
		api.POST("/assets/:id/daily-values", handlers.UpsertAssetDailyValue)

		// Pending settlements
		api.GET("/settlements", handlers.GetSettlements)
		api.POST("/settlements/:id/settle", handlers.SettleOne)
		api.POST("/settlements/run", handlers.RunSettlements)
	}

	r.Run(":8080")
}
