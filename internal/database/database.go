package database

import (
	"database/sql"
	"fmt"
	"time"

	"api-gateway-backend/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

// DB wraps sql.DB
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Item represents an item from external API
type Item struct {
	ID         int64     `json:"id"`
	ExternalID string    `json:"external_id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	UserID     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Order represents an order for analytics queries
type Order struct {
	ID         int64     `json:"id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// OrderStatusSummary represents order summary by status
type OrderStatusSummary struct {
	Status      string  `json:"status"`
	OrderCount  int     `json:"order_count"`
	TotalAmount float64 `json:"total_amount"`
}

// TopCustomer represents top customer by spend
type TopCustomer struct {
	CustomerID  string  `json:"customer_id"`
	TotalSpend  float64 `json:"total_spend"`
	OrderCount  int     `json:"order_count"`
}

// UpsertItem inserts or updates an item (idempotent)
func (db *DB) UpsertItem(item *Item) error {
	query := `
		INSERT INTO items (external_id, title, body, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			body = VALUES(body),
			user_id = VALUES(user_id),
			updated_at = NOW()
	`
	_, err := db.Exec(query, item.ExternalID, item.Title, item.Body, item.UserID)
	return err
}

// GetAllItems retrieves all items from database
func (db *DB) GetAllItems() ([]Item, error) {
	query := `SELECT id, external_id, title, body, user_id, created_at, updated_at FROM items ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.ExternalID, &item.Title, &item.Body, &item.UserID, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetOrderStatusSummary returns order count and total amount by status for last 30 days
func (db *DB) GetOrderStatusSummary() ([]OrderStatusSummary, error) {
	query := `
		SELECT 
			status,
			COUNT(*) as order_count,
			SUM(amount) as total_amount
		FROM orders 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY status
		ORDER BY total_amount DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []OrderStatusSummary
	for rows.Next() {
		var summary OrderStatusSummary
		err := rows.Scan(&summary.Status, &summary.OrderCount, &summary.TotalAmount)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// GetTopCustomers returns top 5 customers by total spend
func (db *DB) GetTopCustomers() ([]TopCustomer, error) {
	query := `
		SELECT 
			customer_id,
			SUM(amount) as total_spend,
			COUNT(*) as order_count
		FROM orders 
		GROUP BY customer_id
		ORDER BY total_spend DESC
		LIMIT 5
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []TopCustomer
	for rows.Next() {
		var customer TopCustomer
		err := rows.Scan(&customer.CustomerID, &customer.TotalSpend, &customer.OrderCount)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, rows.Err()
}