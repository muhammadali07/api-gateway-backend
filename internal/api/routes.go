package api

import (
	"context"
	"net/http"
	"time"

	"api-gateway-backend/internal/config"
	"api-gateway-backend/internal/database"
	"api-gateway-backend/internal/jobs"
	"api-gateway-backend/internal/logger"
	"api-gateway-backend/internal/redis"

	"github.com/gin-gonic/gin"
)

const (
	itemsCacheKey = "items:all"
	itemsCacheTTL = 5 * time.Minute
)

// Handler contains dependencies for API handlers
type Handler struct {
	db        *database.DB
	redis     *redis.Client
	jobManager *jobs.Manager
	logger    *logger.Logger
}

// NewRouter creates a new Gin router with all routes
func NewRouter(db *database.DB, rdb *redis.Client, apiCfg config.ExternalAPIConfig, log *logger.Logger) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Initialize handler
	jobManager := jobs.New(db, rdb, apiCfg, log)
	h := &Handler{
		db:         db,
		redis:      rdb,
		jobManager: jobManager,
		logger:     log,
	}

	// Health check
	router.GET("/health", h.healthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/sync", h.syncData)
		v1.GET("/items", h.getItems)
		v1.GET("/analytics/orders/status", h.getOrderStatusSummary)
		v1.GET("/analytics/customers/top", h.getTopCustomers)
	}

	return router
}

// healthCheck returns the health status of the service
func (h *Handler) healthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check database connection
	if err := h.db.PingContext(ctx); err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	// Check Redis connection
	if err := h.redis.Ping(ctx).Err(); err != nil {
		h.logger.WithError(err).Error("Redis health check failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "redis connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// syncData handles POST /api/v1/sync
func (h *Handler) syncData(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
	defer cancel()

	h.logger.Info("Manual sync requested")

	if err := h.jobManager.SyncDataManual(ctx); err != nil {
		h.logger.WithError(err).Error("Manual sync failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "sync failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "sync completed successfully",
		"timestamp": time.Now().UTC(),
	})
}

// getItems handles GET /api/v1/items with Redis caching
func (h *Handler) getItems(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Try to get from cache first
	var items []database.Item
	if err := h.redis.GetJSON(ctx, itemsCacheKey, &items); err == nil {
		h.logger.Debug("Items served from cache")
		c.Header("X-Cache", "HIT")
		c.JSON(http.StatusOK, gin.H{
			"data":      items,
			"count":     len(items),
			"cached":    true,
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Cache miss - get from database
	items, err := h.db.GetAllItems()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get items from database")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to retrieve items",
			"message": err.Error(),
		})
		return
	}

	// Store in cache for next time
	if err := h.redis.SetJSON(ctx, itemsCacheKey, items, itemsCacheTTL); err != nil {
		h.logger.WithError(err).Warn("Failed to cache items")
	}

	h.logger.WithField("count", len(items)).Debug("Items served from database")
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, gin.H{
		"data":      items,
		"count":     len(items),
		"cached":    false,
		"timestamp": time.Now().UTC(),
	})
}

// getOrderStatusSummary handles GET /api/v1/analytics/orders/status
func (h *Handler) getOrderStatusSummary(c *gin.Context) {
	summaries, err := h.db.GetOrderStatusSummary()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get order status summary")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to retrieve order status summary",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      summaries,
		"timestamp": time.Now().UTC(),
	})
}

// getTopCustomers handles GET /api/v1/analytics/customers/top
func (h *Handler) getTopCustomers(c *gin.Context) {
	customers, err := h.db.GetTopCustomers()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get top customers")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to retrieve top customers",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      customers,
		"timestamp": time.Now().UTC(),
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}