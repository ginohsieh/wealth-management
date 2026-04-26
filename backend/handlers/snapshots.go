package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ginohsieh/wealth-management/backend/store"
)

// GetSnapshots returns all daily net worth snapshots, newest first.
func GetSnapshots(c *gin.Context) {
	snaps := store.GetSnapshots()
	// Sort descending by date in-place (simple insertion sort on small slice)
	for i := 1; i < len(snaps); i++ {
		for j := i; j > 0 && snaps[j].Date > snaps[j-1].Date; j-- {
			snaps[j], snaps[j-1] = snaps[j-1], snaps[j]
		}
	}
	c.JSON(http.StatusOK, snaps)
}

// RecordSnapshot captures the current portfolio / account state for a given date.
// Body: { "date": "YYYY-MM-DD" } — if omitted, today's date is used.
func RecordSnapshot(c *gin.Context) {
	var body struct {
		Date string `json:"date"`
	}
	// Ignore bind error — date is optional
	_ = c.ShouldBindJSON(&body)
	if body.Date == "" {
		body.Date = time.Now().Format("2006-01-02")
	}
	snap := store.RecordSnapshot(body.Date)
	c.JSON(http.StatusCreated, snap)
}
