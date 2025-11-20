#!/bin/bash

# TrainBlink Server Testing Script
# Tests all HTTP and WebSocket endpoints

set -e

BASE_URL="http://localhost:8080"
WS_URL="ws://localhost:8080/ws"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "╔══════════════════════════════════════════════════════╗"
echo "║                                                      ║"
echo "║     🚄 TrainBlink Server Test Suite                 ║"
echo "║                                                      ║"
echo "╚══════════════════════════════════════════════════════╝"
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test endpoints
test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    local data=$5

    echo -n "Testing $description... "

    if [ "$method" == "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi

    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$status_code" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (Status: $status_code)"
        echo "   Response: $(echo $body | head -c 100)..."
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC} (Expected: $expected_status, Got: $status_code)"
        echo "   Response: $body"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    echo ""
}

# Check if server is running
echo "Checking if server is running..."
if ! curl -s "$BASE_URL/ping" > /dev/null 2>&1; then
    echo -e "${RED}❌ Server is not running!${NC}"
    echo "Please start the server first:"
    echo "  go run cmd/server/main_websocket.go"
    exit 1
fi
echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# Test 1: Root endpoint
test_endpoint "GET" "/" "200" "Root endpoint"

# Test 2: Ping
test_endpoint "GET" "/ping" "200" "Ping endpoint"

# Test 3: Health
test_endpoint "GET" "/health" "200" "Health check"

# Test 4: Hello
test_endpoint "GET" "/api/v1/hello" "200" "Hello endpoint"

# Test 5: Welcome with parameters
test_endpoint "GET" "/api/v1/welcome?name=Alice&device_id=test-123" "200" "Welcome endpoint"

# Test 6: Connections
test_endpoint "GET" "/api/v1/connections" "200" "Connections list"

# Test 7: Send message (broadcast)
test_endpoint "POST" "/api/v1/message" "200" "Send broadcast message" \
    '{"from":"test-client","to":"all","message":"Hello from test!"}'

# Test 8: Send message (specific client)
test_endpoint "POST" "/api/v1/message" "404" "Send to non-existent client" \
    '{"from":"test-client","to":"client-999","message":"Hello!"}'

echo ""
echo "════════════════════════════════════════════════════════"
echo ""
echo "Test Results:"
echo -e "  ${GREEN}✓ Passed: $TESTS_PASSED${NC}"
echo -e "  ${RED}✗ Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
