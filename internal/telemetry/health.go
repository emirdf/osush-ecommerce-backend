package telemetry

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type HealthChecker struct {
	db *pgxpool.Pool
}

func NewHealthChecker(db *pgxpool.Pool) *HealthChecker {
	return &HealthChecker{
		db: db,
	}
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (h *HealthChecker) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	checks := map[string]CheckResult{
		"database": h.checkDatabase(ctx),
	}

	overallStatus := "UP"

	for _, check := range checks {
		if check.Status != "UP" {
			overallStatus = "DOWN"
			break
		}
	}

	status := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	if overallStatus == "DOWN" {
		c.JSON(503, status)
		return
	}

	c.JSON(200, status)
}

func (h *HealthChecker) checkDatabase(ctx context.Context) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return CheckResult{
			Status:  "DOWN",
			Message: err.Error(),
		}
	}

	return CheckResult{
		Status: "UP",
	}
}

func (h *HealthChecker) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		2*time.Second,
	)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		c.JSON(503, gin.H{
			"status": "NOT_READY",
			"error":  "database not ready",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "READY",
	})
}

func (h *HealthChecker) LivenessCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ALIVE",
	})
}
