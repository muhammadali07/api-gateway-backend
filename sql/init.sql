-- Create database if not exists
CREATE DATABASE IF NOT EXISTS api_gateway;
USE api_gateway;

-- Items table for storing external API data
CREATE TABLE IF NOT EXISTS items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    external_id VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(500) NOT NULL,
    body TEXT,
    user_id INT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_external_id (external_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- Orders table for Part 3 SQL queries
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status ENUM('PENDING', 'PAID', 'CANCELLED') NOT NULL,
    created_at DATETIME NOT NULL,
    INDEX idx_customer_id (customer_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Insert sample orders data for testing
INSERT INTO orders (customer_id, amount, status, created_at) VALUES
('customer-1', 100.50, 'PAID', DATE_SUB(NOW(), INTERVAL 5 DAY)),
('customer-2', 250.75, 'PENDING', DATE_SUB(NOW(), INTERVAL 10 DAY)),
('customer-1', 75.25, 'PAID', DATE_SUB(NOW(), INTERVAL 15 DAY)),
('customer-3', 500.00, 'CANCELLED', DATE_SUB(NOW(), INTERVAL 20 DAY)),
('customer-2', 150.00, 'PAID', DATE_SUB(NOW(), INTERVAL 25 DAY)),
('customer-4', 300.25, 'PAID', DATE_SUB(NOW(), INTERVAL 3 DAY)),
('customer-1', 125.50, 'PENDING', DATE_SUB(NOW(), INTERVAL 7 DAY)),
('customer-5', 200.00, 'PAID', DATE_SUB(NOW(), INTERVAL 12 DAY)),
('customer-3', 450.75, 'PAID', DATE_SUB(NOW(), INTERVAL 18 DAY)),
('customer-2', 175.25, 'CANCELLED', DATE_SUB(NOW(), INTERVAL 22 DAY));