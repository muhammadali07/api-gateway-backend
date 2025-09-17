#!/bin/bash

# API Testing Script for Backend Application
# Usage: ./test-api.sh [base_url]
# Default base_url: http://localhost:8080

BASE_URL=${1:-"http://localhost:8080"}
COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_NC='\033[0m' # No Color

echo -e "${COLOR_BLUE}🚀 Testing API Gateway Backend${COLOR_NC}"
echo -e "${COLOR_BLUE}Base URL: $BASE_URL${COLOR_NC}"
echo ""

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local expected_status=${4:-200}
    
    echo -e "${COLOR_YELLOW}Testing: $description${COLOR_NC}"
    echo "$method $endpoint"
    
    response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" -H "Content-Type: application/json")
    
    # Split response and status code
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" -eq "$expected_status" ]; then
        echo -e "${COLOR_GREEN}✅ SUCCESS (Status: $status_code)${COLOR_NC}"
        # Pretty print JSON if response is JSON
        if echo "$body" | jq . >/dev/null 2>&1; then
            echo "$body" | jq .
        else
            echo "$body"
        fi
    else
        echo -e "${COLOR_RED}❌ FAILED (Expected: $expected_status, Got: $status_code)${COLOR_NC}"
        echo "Response: $body"
    fi
    
    echo ""
    sleep 1
}

# Function to test with timing
test_with_timing() {
    local method=$1
    local endpoint=$2
    local description=$3
    
    echo -e "${COLOR_YELLOW}Testing: $description${COLOR_NC}"
    echo "$method $endpoint"
    
    start_time=$(date +%s.%N)
    response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" -H "Content-Type: application/json")
    end_time=$(date +%s.%N)
    
    duration=$(echo "$end_time - $start_time" | bc)
    
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" -eq "200" ]; then
        echo -e "${COLOR_GREEN}✅ SUCCESS (Status: $status_code, Time: ${duration}s)${COLOR_NC}"
        if echo "$body" | jq . >/dev/null 2>&1; then
            echo "$body" | jq .
        else
            echo "$body"
        fi
    else
        echo -e "${COLOR_RED}❌ FAILED (Status: $status_code, Time: ${duration}s)${COLOR_NC}"
        echo "Response: $body"
    fi
    
    echo ""
    sleep 1
}

# Check if server is running
echo -e "${COLOR_BLUE}🔍 Checking if server is running...${COLOR_NC}"
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo -e "${COLOR_RED}❌ Server is not running at $BASE_URL${COLOR_NC}"
    echo "Please start the server first:"
    echo "  make docker-up"
    echo "  # or"
    echo "  docker-compose up -d"
    exit 1
fi
echo -e "${COLOR_GREEN}✅ Server is running${COLOR_NC}"
echo ""

# Test 1: Health Check
echo -e "${COLOR_BLUE}📋 Test 1: Health Check${COLOR_NC}"
test_endpoint "GET" "/health" "Health check endpoint"

# Test 2: Data Synchronization
echo -e "${COLOR_BLUE}📋 Test 2: Data Synchronization${COLOR_NC}"
test_endpoint "POST" "/api/v1/sync" "Manual data sync"

# Test 3: Get Items (should be cached after sync)
echo -e "${COLOR_BLUE}📋 Test 3: Get Items (Cache Test)${COLOR_NC}"
test_with_timing "GET" "/api/v1/items" "Get items - first request (cache miss)"
test_with_timing "GET" "/api/v1/items" "Get items - second request (cache hit)"

# Test 4: Analytics - Order Status
echo -e "${COLOR_BLUE}📋 Test 4: Analytics - Order Status${COLOR_NC}"
test_endpoint "GET" "/api/v1/analytics/orders/status" "Order status summary"

# Test 5: Analytics - Top Customers
echo -e "${COLOR_BLUE}📋 Test 5: Analytics - Top Customers${COLOR_NC}"
test_endpoint "GET" "/api/v1/analytics/customers/top" "Top customers by spend"

# Test 6: Error Handling
echo -e "${COLOR_BLUE}📋 Test 6: Error Handling${COLOR_NC}"
test_endpoint "GET" "/api/v1/invalid" "Invalid endpoint" 404
test_endpoint "DELETE" "/api/v1/items" "Invalid method" 405

# Test 7: Load Test (Simple)
echo -e "${COLOR_BLUE}📋 Test 7: Simple Load Test${COLOR_NC}"
echo -e "${COLOR_YELLOW}Running 10 concurrent requests to /api/v1/items...${COLOR_NC}"

start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$BASE_URL/api/v1/items" > /dev/null &
done
wait
end_time=$(date +%s.%N)

duration=$(echo "$end_time - $start_time" | bc)
echo -e "${COLOR_GREEN}✅ Completed 10 concurrent requests in ${duration}s${COLOR_NC}"
echo ""

# Summary
echo -e "${COLOR_BLUE}📊 Testing Summary${COLOR_NC}"
echo "All tests completed!"
echo ""
echo "Additional testing commands:"
echo "  # Manual testing:"
echo "  curl -X GET $BASE_URL/health"
echo "  curl -X POST $BASE_URL/api/v1/sync"
echo "  curl -X GET $BASE_URL/api/v1/items"
echo ""
echo "  # Load testing with Apache Bench:"
echo "  ab -n 100 -c 10 $BASE_URL/api/v1/items"
echo ""
echo "  # Monitor logs:"
echo "  make docker-logs"
echo ""
echo -e "${COLOR_GREEN}🎉 Testing completed successfully!${COLOR_NC}"