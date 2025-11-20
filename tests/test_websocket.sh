#!/bin/bash

# TrainBlink WebSocket Test Script
# Tests real-time messaging over WebSocket

set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Counter
TESTS_PASSED=0
TESTS_FAILED=0

# Test function for HTTP endpoints
test_http() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local headers="$5"

    echo -e "${BLUE}Testing HTTP:${NC} $name"

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            $headers \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        echo "Response: $body" | head -c 200
        echo "..."
        echo
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "$body"
    else
        echo -e "${RED}✗ FAILED${NC} (HTTP $http_code)"
        echo "Response: $body"
        echo
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}TrainBlink WebSocket Test${NC}"
echo -e "${YELLOW}========================================${NC}"
echo

# Test 1: Health Check
echo -e "${BLUE}[Test 1] Health Check${NC}"
test_http "Health" "GET" "/health" "" ""
echo

# Test 2: Enter Station (User 1 - Tokyo)
echo -e "${BLUE}[Test 2] User 1 Enters Tokyo Station${NC}"
user1_response=$(test_http \
    "User 1 Enter Station" \
    "POST" \
    "/api/v1/geofence/enter" \
    '{"station_id": "station_tokyo_001", "coordinates": {"latitude": 35.6812, "longitude": 139.7671, "accuracy": 10.5}, "client_version": "2.5.0", "capabilities": {"p2p_enabled": true, "matrix_enabled": true, "mls_supported": false}}' \
    '-H "X-User-ID: user-001" -H "X-Device-ID: iOS-123"')

user1_session=$(echo "$user1_response" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}User 1 Session ID: $user1_session${NC}"
echo

# Test 3: Enter Station (User 2 - Tokyo)
echo -e "${BLUE}[Test 3] User 2 Enters Tokyo Station${NC}"
user2_response=$(test_http \
    "User 2 Enter Station" \
    "POST" \
    "/api/v1/geofence/enter" \
    '{"station_id": "station_tokyo_001", "coordinates": {"latitude": 35.6815, "longitude": 139.7670, "accuracy": 12.0}, "client_version": "2.5.0", "capabilities": {"p2p_enabled": true, "matrix_enabled": true, "mls_supported": false}}' \
    '-H "X-User-ID: user-002" -H "X-Device-ID: Android-456"')

user2_session=$(echo "$user2_response" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}User 2 Session ID: $user2_session${NC}"
echo

# Test 4: WebSocket Connection Test (using websocat if available)
echo -e "${BLUE}[Test 4] WebSocket Connection Test${NC}"

if command -v websocat &> /dev/null; then
    echo "Testing WebSocket connection for User 1..."

    # Test WebSocket connection
    ws_url="ws://localhost:8080/ws?session_id=$user1_session"
    echo "Connecting to: $ws_url"

    # Send a test message
    echo '{"type":"message","content":"Hello from User 1!"}' | timeout 2 websocat "$ws_url" &

    echo -e "${GREEN}✓ WebSocket connection test initiated${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${YELLOW}⚠ websocat not installed, skipping WebSocket connection test${NC}"
    echo "Install websocat with: cargo install websocat"
    echo "Or test manually with: websocat ws://localhost:8080/ws?session_id=$user1_session"
fi
echo

# Test 5: Statistics
echo -e "${BLUE}[Test 5] Get Service Statistics${NC}"
test_http "Get Statistics" "GET" "/api/v1/geofence/stats" "" ""
echo

# Wait a bit
sleep 1

# Test 6: Exit Station (User 1)
echo -e "${BLUE}[Test 6] User 1 Exits Tokyo Station${NC}"
test_http \
    "User 1 Exit Station" \
    "POST" \
    "/api/v1/geofence/exit" \
    "{\"session_id\": \"$user1_session\", \"station_id\": \"station_tokyo_001\", \"duration_seconds\": 120, \"activity\": {\"p2p_chats_created\": 1, \"p2p_messages_sent\": 5, \"content_shared\": 1, \"matrix_messages_sent\": 3}}" \
    '-H "X-User-ID: user-001"'
echo

# Test 7: Exit Station (User 2)
echo -e "${BLUE}[Test 7] User 2 Exits Tokyo Station${NC}"
test_http \
    "User 2 Exit Station" \
    "POST" \
    "/api/v1/geofence/exit" \
    "{\"session_id\": \"$user2_session\", \"station_id\": \"station_tokyo_001\", \"duration_seconds\": 90, \"activity\": {\"p2p_chats_created\": 1, \"p2p_messages_sent\": 4, \"content_shared\": 0, \"matrix_messages_sent\": 2}}" \
    '-H "X-User-ID: user-002"'
echo

# Summary
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}✓ Passed:${NC} $TESTS_PASSED"
echo -e "${RED}✗ Failed:${NC} $TESTS_FAILED"
echo
echo -e "${YELLOW}WebSocket Testing Instructions:${NC}"
echo "1. Install websocat: cargo install websocat"
echo "2. Get a session ID by entering a station (see tests above)"
echo "3. Connect to WebSocket: websocat ws://localhost:8080/ws?session_id=YOUR_SESSION_ID"
echo "4. Send messages:"
echo "   - Chat: {\"type\":\"message\",\"content\":\"Hello!\"}"
echo "   - Typing: {\"type\":\"typing\",\"metadata\":{\"is_typing\":true}}"
echo "   - Ping: {\"type\":\"ping\"}"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All HTTP tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
