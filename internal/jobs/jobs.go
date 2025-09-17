package jobs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"api-gateway-backend/internal/client"
	"api-gateway-backend/internal/config"
	"api-gateway-backend/internal/database"
	"api-gateway-backend/internal/logger"
	"api-gateway-backend/internal/redis"

	"github.com/robfig/cron/v3"
)

// Manager handles background jobs
type Manager struct {
	cron      *cron.Cron
	db        *database.DB
	redis     *redis.Client
	client    *client.ExternalAPIClient
	logger    *logger.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

// New creates a new job manager
func New(db *database.DB, rdb *redis.Client, apiCfg config.ExternalAPIConfig, log *logger.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		cron:   cron.New(cron.WithSeconds()),
		db:     db,
		redis:  rdb,
		client: client.New(apiCfg),
		logger: log,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the background jobs
func (m *Manager) Start() {
	// Schedule data sync every 15 minutes
	_, err := m.cron.AddFunc("0 */15 * * * *", func() {
		if err := m.syncData(); err != nil {
			m.logger.WithError(err).Error("Failed to sync data")
		}
	})
	if err != nil {
		m.logger.WithError(err).Error("Failed to schedule sync job")
		return
	}

	m.cron.Start()
	m.logger.Info("Background jobs started")

	// Run initial sync
	go func() {
		if err := m.syncData(); err != nil {
			m.logger.WithError(err).Error("Failed to perform initial sync")
		}
	}()
}

// Stop stops the background jobs
func (m *Manager) Stop() {
	m.cancel()
	m.cron.Stop()
	m.logger.Info("Background jobs stopped")
}

// syncData fetches data from external API and stores in database
func (m *Manager) syncData() error {
	ctx, cancel := context.WithTimeout(m.ctx, 2*time.Minute)
	defer cancel()

	m.logger.Info("Starting data sync")
	start := time.Now()

	// Fetch posts from external API
	posts, err := m.client.FetchPosts(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %w", err)
	}

	m.logger.WithField("count", len(posts)).Info("Fetched posts from external API")

	// Store posts in database (idempotent)
	var successCount, errorCount int
	for _, post := range posts {
		item := &database.Item{
			ExternalID: strconv.Itoa(post.ID),
			Title:      post.Title,
			Body:       post.Body,
			UserID:     post.UserID,
		}

		if err := m.db.UpsertItem(item); err != nil {
			m.logger.WithError(err).WithField("external_id", item.ExternalID).Error("Failed to upsert item")
			errorCount++
		} else {
			successCount++
		}
	}

	// Invalidate cache after successful sync
	if successCount > 0 {
		if err := m.redis.InvalidatePattern(ctx, "items:*"); err != nil {
			m.logger.WithError(err).Warn("Failed to invalidate cache")
		}
	}

	duration := time.Since(start)
	m.logger.WithFields(map[string]interface{}{
		"success_count": successCount,
		"error_count":   errorCount,
		"duration":      duration,
	}).Info("Data sync completed")

	if errorCount > 0 {
		return fmt.Errorf("sync completed with %d errors out of %d items", errorCount, len(posts))
	}

	return nil
}

// SyncDataManual performs manual data sync (for /sync endpoint)
func (m *Manager) SyncDataManual(ctx context.Context) error {
	return m.syncData()
}