# API Testing Payloads

Berikut adalah contoh payload dan cara testing untuk semua endpoint API yang tersedia.

## 🔍 Health Check

### GET /health

**Request:**
```bash
curl -X GET http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "database": "connected",
  "redis": "connected"
}
```

## 🔄 Data Synchronization

### POST /api/v1/sync

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/sync \
  -H "Content-Type: application/json"
```

**Expected Response (Success):**
```json
{
  "message": "Data synchronized successfully",
  "items_processed": 100,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Expected Response (Error):**
```json
{
  "error": "Failed to fetch data from external API",
  "details": "connection timeout"
}
```

## 📦 Items Endpoint

### GET /api/v1/items

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/items
```

**Expected Response:**
```json
{
  "items": [
    {
      "id": 1,
      "external_id": "1",
      "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
      "body": "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto",
      "user_id": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "external_id": "2",
      "title": "qui est esse",
      "body": "est rerum tempore vitae\nsequi sint nihil reprehenderit dolor beatae ea dolores neque\nfugiat blanditiis voluptate porro vel nihil molestiae ut reiciendis\nqui aperiam non debitis possimus qui neque nisi nulla",
      "user_id": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 100,
  "cached": true,
  "cache_expires_at": "2024-01-15T10:35:00Z"
}
```

## 📊 Analytics Endpoints

### GET /api/v1/analytics/orders/status

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/analytics/orders/status
```

**Expected Response:**
```json
{
  "data": [
    {
      "status": "PAID",
      "order_count": 45,
      "total_amount": 12500.75
    },
    {
      "status": "PENDING",
      "order_count": 23,
      "total_amount": 6750.25
    },
    {
      "status": "CANCELLED",
      "order_count": 8,
      "total_amount": 2100.00
    }
  ],
  "period": "last_30_days",
  "generated_at": "2024-01-15T10:30:00Z"
}
```

### GET /api/v1/analytics/customers/top

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/analytics/customers/top
```

**Expected Response:**
```json
{
  "data": [
    {
      "customer_id": "customer-001",
      "total_spend": 5500.75,
      "order_count": 12
    },
    {
      "customer_id": "customer-002",
      "total_spend": 4200.50,
      "order_count": 8
    },
    {
      "customer_id": "customer-003",
      "total_spend": 3800.25,
      "order_count": 15
    },
    {
      "customer_id": "customer-004",
      "total_spend": 3200.00,
      "order_count": 6
    },
    {
      "customer_id": "customer-005",
      "total_spend": 2900.75,
      "order_count": 9
    }
  ],
  "limit": 5,
  "generated_at": "2024-01-15T10:30:00Z"
}
```

## 🧪 Testing Scenarios

### 1. Complete Flow Testing

```bash
# 1. Check health
curl -X GET http://localhost:8080/health

# 2. Sync data
curl -X POST http://localhost:8080/api/v1/sync

# 3. Get items (should be cached)
curl -X GET http://localhost:8080/api/v1/items

# 4. Get analytics
curl -X GET http://localhost:8080/api/v1/analytics/orders/status
curl -X GET http://localhost:8080/api/v1/analytics/customers/top
```

### 2. Error Handling Testing

```bash
# Test invalid endpoint
curl -X GET http://localhost:8080/api/v1/invalid
# Expected: 404 Not Found

# Test invalid method
curl -X DELETE http://localhost:8080/api/v1/items
# Expected: 405 Method Not Allowed
```

### 3. Cache Testing

```bash
# First request (cache miss)
curl -X GET http://localhost:8080/api/v1/items
# Check response time and "cached": false

# Second request (cache hit)
curl -X GET http://localhost:8080/api/v1/items
# Should be faster and "cached": true

# Sync data (invalidates cache)
curl -X POST http://localhost:8080/api/v1/sync

# Request items again (cache miss after sync)
curl -X GET http://localhost:8080/api/v1/items
# "cached": false again
```

## 🔧 Using Postman Collection

Jika menggunakan Postman, buat collection dengan:

### Environment Variables
```json
{
  "base_url": "http://localhost:8080"
}
```

### Collection Structure
```
API Gateway Backend
├── Health Check
│   └── GET {{base_url}}/health
├── Data Sync
│   └── POST {{base_url}}/api/v1/sync
├── Items
│   └── GET {{base_url}}/api/v1/items
└── Analytics
    ├── GET {{base_url}}/api/v1/analytics/orders/status
    └── GET {{base_url}}/api/v1/analytics/customers/top
```

## 🐛 Common Error Responses

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "message": "Database connection failed"
}
```

### 503 Service Unavailable
```json
{
  "error": "Service temporarily unavailable",
  "message": "External API is down"
}
```

### 404 Not Found
```json
{
  "error": "Not found",
  "message": "Endpoint not found"
}
```

## 📈 Performance Testing

### Load Testing dengan curl
```bash
# Test concurrent requests
for i in {1..10}; do
  curl -X GET http://localhost:8080/api/v1/items &
done
wait
```

### Using Apache Bench (ab)
```bash
# Test 100 requests with 10 concurrent connections
ab -n 100 -c 10 http://localhost:8080/api/v1/items
```

## 🔍 Monitoring & Debugging

### Check Logs
```bash
# View application logs
make docker-logs

# Or specific service logs
docker-compose logs -f api-gateway
```

### Check Service Status
```bash
# Check all services
make monitor

# Check specific service
docker-compose ps
```

### Redis Cache Inspection
```bash
# Connect to Redis
docker-compose exec redis redis-cli

# Check cached keys
KEYS items:*

# Get cached data
GET items:all

# Check TTL
TTL items:all
```

### Database Inspection
```bash
# Connect to MySQL
docker-compose exec mysql mysql -u apiuser -p api_gateway

# Check items table
SELECT COUNT(*) FROM items;
SELECT * FROM items LIMIT 5;

# Check orders table
SELECT COUNT(*) FROM orders;
SELECT * FROM orders LIMIT 5;
```