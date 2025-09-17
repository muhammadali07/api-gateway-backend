package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-gateway-backend/internal/config"
	"api-gateway-backend/internal/database"
	"api-gateway-backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock implementation of database operations
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) PingContext(ctx interface{}) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDB) UpsertItem(item *database.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockDB) GetAllItems() ([]database.Item, error) {
	args := m.Called()
	return args.Get(0).([]database.Item), args.Error(1)
}

func (m *MockDB) GetOrderStatusSummary() ([]database.OrderStatusSummary, error) {
	args := m.Called()
	return args.Get(0).([]database.OrderStatusSummary), args.Error(1)
}

func (m *MockDB) GetTopCustomers() ([]database.TopCustomer, error) {
	args := m.Called()
	return args.Get(0).([]database.TopCustomer), args.Error(1)
}

// MockRedis is a mock implementation of Redis operations
type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedis) Ping(ctx interface{}) *MockResult {
	args := m.Called(ctx)
	return args.Get(0).(*MockResult)
}

func (m *MockRedis) GetJSON(ctx interface{}, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockRedis) SetJSON(ctx interface{}, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockRedis) InvalidatePattern(ctx interface{}, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// MockResult is a mock Redis result
type MockResult struct {
	mock.Mock
}

func (m *MockResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockJobManager is a mock implementation of job manager
type MockJobManager struct {
	mock.Mock
}

func (m *MockJobManager) SyncDataManual(ctx interface{}) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupTestRouter() (*gin.Engine, *MockDB, *MockRedis, *MockJobManager) {
	gin.SetMode(gin.TestMode)

	mockDB := &MockDB{}
	mockRedis := &MockRedis{}
	mockJobManager := &MockJobManager{}
	logger := logger.New()

	router := gin.New()
	h := &Handler{
		db:         mockDB,
		redis:      mockRedis,
		jobManager: mockJobManager,
		logger:     logger,
	}

	// Add routes
	router.GET("/health", h.healthCheck)
	v1 := router.Group("/api/v1")
	{
		v1.POST("/sync", h.syncData)
		v1.GET("/items", h.getItems)
		v1.GET("/analytics/orders/status", h.getOrderStatusSummary)
		v1.GET("/analytics/customers/top", h.getTopCustomers)
	}

	return router, mockDB, mockRedis, mockJobManager
}

func TestHealthCheck_Success(t *testing.T) {
	router, mockDB, mockRedis, _ := setupTestRouter()

	// Setup mocks
	mockResult := &MockResult{}
	mockResult.On("Err").Return(nil)
	mockDB.On("PingContext", mock.Anything).Return(nil)
	mockRedis.On("Ping", mock.Anything).Return(mockResult)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestHealthCheck_DatabaseFailure(t *testing.T) {
	router, mockDB, _, _ := setupTestRouter()

	// Setup mocks - database fails
	mockDB.On("PingContext", mock.Anything).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unhealthy", response["status"])
	assert.Equal(t, "database connection failed", response["error"])

	mockDB.AssertExpectations(t)
}

func TestSyncData_Success(t *testing.T) {
	router, _, _, mockJobManager := setupTestRouter()

	// Setup mocks
	mockJobManager.On("SyncDataManual", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/sync", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "sync completed successfully", response["message"])

	mockJobManager.AssertExpectations(t)
}

func TestSyncData_Failure(t *testing.T) {
	router, _, _, mockJobManager := setupTestRouter()

	// Setup mocks - sync fails
	mockJobManager.On("SyncDataManual", mock.Anything).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/sync", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "sync failed", response["error"])

	mockJobManager.AssertExpectations(t)
}

func TestGetItems_CacheHit(t *testing.T) {
	router, _, mockRedis, _ := setupTestRouter()

	// Setup mocks - cache hit
	expectedItems := []database.Item{
		{ID: 1, ExternalID: "1", Title: "Test Item", Body: "Test Body", UserID: 1},
	}
	mockRedis.On("GetJSON", mock.Anything, "items:all", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*[]database.Item)
		*dest = expectedItems
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "HIT", w.Header().Get("X-Cache"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["cached"])
	assert.Equal(t, float64(1), response["count"])

	mockRedis.AssertExpectations(t)
}

func TestGetItems_CacheMiss(t *testing.T) {
	router, mockDB, mockRedis, _ := setupTestRouter()

	// Setup mocks - cache miss, database hit
	expectedItems := []database.Item{
		{ID: 1, ExternalID: "1", Title: "Test Item", Body: "Test Body", UserID: 1},
	}
	mockRedis.On("GetJSON", mock.Anything, "items:all", mock.Anything).Return(assert.AnError)
	mockDB.On("GetAllItems").Return(expectedItems, nil)
	mockRedis.On("SetJSON", mock.Anything, "items:all", expectedItems, itemsCacheTTL).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "MISS", w.Header().Get("X-Cache"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["cached"])
	assert.Equal(t, float64(1), response["count"])

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestGetOrderStatusSummary_Success(t *testing.T) {
	router, mockDB, _, _ := setupTestRouter()

	// Setup mocks
	expectedSummaries := []database.OrderStatusSummary{
		{Status: "PAID", OrderCount: 10, TotalAmount: 1500.50},
		{Status: "PENDING", OrderCount: 5, TotalAmount: 750.25},
	}
	mockDB.On("GetOrderStatusSummary").Return(expectedSummaries, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/orders/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	mockDB.AssertExpectations(t)
}

func TestGetTopCustomers_Success(t *testing.T) {
	router, mockDB, _, _ := setupTestRouter()

	// Setup mocks
	expectedCustomers := []database.TopCustomer{
		{CustomerID: "customer-1", TotalSpend: 2500.75, OrderCount: 15},
		{CustomerID: "customer-2", TotalSpend: 1800.25, OrderCount: 12},
	}
	mockDB.On("GetTopCustomers").Return(expectedCustomers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/customers/top", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	mockDB.AssertExpectations(t)
}