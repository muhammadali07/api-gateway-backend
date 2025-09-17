# API Gateway Backend

A robust Go backend service that integrates with external APIs, featuring MySQL storage, Redis caching, background job processing, and comprehensive analytics. Built for the Backend Engineer Take-Home Test.

## 🚀 Features

- **External API Integration**: Fetches data from JSONPlaceholder API with retry logic and error handling
- **Database Operations**: MySQL with idempotent writes and optimized queries
- **Redis Caching**: Intelligent caching with TTL and invalidation strategies
- **Background Jobs**: Automated data synchronization every 15 minutes
- **Analytics Endpoints**: Order status summaries and top customer analytics
- **Production Ready**: Comprehensive logging, monitoring, health checks, and graceful shutdown
- **Docker Support**: Full containerization with Docker Compose
- **Unit Testing**: Comprehensive test coverage with mocks

## 📋 API Endpoints

### Core Endpoints
- `GET /health` - Health check endpoint
- `POST /api/v1/sync` - Manual data synchronization
- `GET /api/v1/items` - Retrieve cached items

### Analytics Endpoints
- `GET /api/v1/analytics/orders/status` - Order count and total amount by status (last 30 days)
- `GET /api/v1/analytics/customers/top` - Top 5 customers by total spend

## 🛠 Tech Stack

- **Language**: Go 1.21
- **Database**: MySQL 8.0
- **Cache**: Redis 7
- **Web Framework**: Gin
- **Job Scheduler**: Cron v3
- **Testing**: Testify with mocks
- **Containerization**: Docker & Docker Compose

## 🏃‍♂️ How to Run

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- Make (optional, for convenience commands)

### Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd api-gateway-backend
   ```

2. **Start all services**
   ```bash
   make docker-up
   # or
   docker-compose up -d
   ```

3. **Verify the service is running**
   ```bash
   curl http://localhost:8080/health
   ```

4. **Test the API endpoints**
   ```bash
   make api-test
   # or manually:
   curl -X POST http://localhost:8080/api/v1/sync
   curl http://localhost:8080/api/v1/items
   ```

### Local Development

1. **Set up the development environment**
   ```bash
   make dev-setup
   ```

2. **Start dependencies (MySQL & Redis)**
   ```bash
   docker-compose up -d mysql redis
   ```

3. **Run the application locally**
   ```bash
   make run
   # or
   go run ./cmd/server
   ```

### Available Make Commands

```bash
make help                 # Show all available commands
make build               # Build the application
make test                # Run tests with coverage
make docker-up           # Start all services
make docker-down         # Stop all services
make docker-logs         # View service logs
make api-test           # Test API endpoints
make monitor            # Show service status
```

## 🏗 Project Structure

```
api-gateway-backend/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── api/            # HTTP handlers and routes
│   ├── client/         # External API client
│   ├── config/         # Configuration management
│   ├── database/       # Database operations
│   ├── jobs/           # Background job processing
│   ├── logger/         # Logging utilities
│   └── redis/          # Redis operations
├── sql/                # Database initialization
├── docker-compose.yml  # Service orchestration
├── Dockerfile         # Application container
├── Makefile          # Development commands
└── go.mod            # Go dependencies
```

## 🔧 Configuration

The application uses environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DB_HOST` | `localhost` | MySQL host |
| `DB_PORT` | `3306` | MySQL port |
| `DB_USER` | `apiuser` | MySQL username |
| `DB_PASSWORD` | `apipassword` | MySQL password |
| `DB_NAME` | `api_gateway` | MySQL database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `EXTERNAL_API_URL` | `https://jsonplaceholder.typicode.com` | External API base URL |
| `LOG_LEVEL` | `info` | Logging level |
| `ENVIRONMENT` | `development` | Application environment |

## 📊 Database Schema

### Items Table
```sql
CREATE TABLE items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    external_id VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(500) NOT NULL,
    body TEXT,
    user_id INT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Orders Table (for analytics)
```sql
CREATE TABLE orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status ENUM('PENDING', 'PAID', 'CANCELLED') NOT NULL,
    created_at DATETIME NOT NULL
);
```

## 🧪 Testing

### Run Tests
```bash
make test                # Run tests with coverage report
make test-short         # Run tests without coverage
```

### Test Coverage
The project includes comprehensive unit tests with mocks for:
- API handlers
- Database operations
- Redis caching
- External API client
- Background jobs

## 📈 Monitoring & Logging

### Health Checks
- Database connectivity
- Redis connectivity
- Service status

### Logging
- Structured logging with logrus
- Configurable log levels
- JSON format in production
- Request/response logging

### Monitoring
```bash
make monitor            # View service status and resource usage
make docker-logs        # View application logs
```

## 🔄 Background Jobs

- **Data Sync**: Runs every 15 minutes
- **Idempotent Operations**: Prevents duplicate data
- **Error Handling**: Retry logic with exponential backoff
- **Cache Invalidation**: Automatic cache clearing after sync

## 🎯 Key Design Decisions

### Assumptions Made

1. **External API**: Using JSONPlaceholder API as it's free, reliable, and provides structured data
2. **Data Model**: Posts from the API are stored as "items" with external_id for idempotency
3. **Caching Strategy**: 5-minute TTL for items cache with pattern-based invalidation
4. **Background Jobs**: 15-minute interval balances freshness with API rate limits
5. **Error Handling**: Graceful degradation with proper HTTP status codes
6. **Database**: MySQL chosen for ACID compliance and complex query support

### Trade-offs & Improvements

#### Current Implementation
- ✅ Simple and reliable
- ✅ Good separation of concerns
- ✅ Comprehensive error handling
- ✅ Production-ready logging

#### With More Time, I Would Add:

1. **Enhanced Monitoring**
   - Prometheus metrics
   - Grafana dashboards
   - Alert manager integration
   - Distributed tracing with Jaeger

2. **Security Improvements**
   - API authentication/authorization
   - Rate limiting middleware
   - Input validation and sanitization
   - TLS/HTTPS configuration

3. **Performance Optimizations**
   - Database connection pooling tuning
   - Redis clustering for high availability
   - API response pagination
   - Database query optimization

4. **Operational Excellence**
   - Database migrations system
   - Configuration management (Vault/Consul)
   - Blue-green deployment support
   - Automated backup strategies

5. **Advanced Features**
   - WebSocket support for real-time updates
   - Event-driven architecture with message queues
   - Multi-tenant support
   - API versioning strategy

6. **Testing Enhancements**
   - Integration tests with test containers
   - Load testing with k6
   - Contract testing
   - Chaos engineering tests

## 🚨 Troubleshooting

### Common Issues

1. **Service won't start**
   ```bash
   make docker-logs        # Check logs
   docker-compose ps       # Check service status
   ```

2. **Database connection issues**
   ```bash
   docker-compose exec mysql mysql -u apiuser -p api_gateway
   ```

3. **Redis connection issues**
   ```bash
   docker-compose exec redis redis-cli ping
   ```

4. **External API issues**
   - Check internet connectivity
   - Verify API endpoint availability
   - Review retry logic in logs

### Reset Environment
```bash
make dev-reset          # Complete environment reset
```

## 📝 SQL Queries (Part 3)

The application includes the requested SQL queries:

### 1. Orders and Total Amount by Status (Last 30 Days)
```sql
SELECT 
    status,
    COUNT(*) as order_count,
    SUM(amount) as total_amount
FROM orders 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY status
ORDER BY total_amount DESC;
```

### 2. Top 5 Customers by Total Spend
```sql
SELECT 
    customer_id,
    SUM(amount) as total_spend,
    COUNT(*) as order_count
FROM orders 
GROUP BY customer_id
ORDER BY total_spend DESC
LIMIT 5;
```

These queries are implemented in the database layer and exposed via REST endpoints.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is created for the Backend Engineer Take-Home Test.
