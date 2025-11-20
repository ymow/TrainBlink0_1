#!/bin/bash

# TrainBlink Geofence + Matrix Hybrid Test Script
# Tests the complete hybrid architecture (P2P + Matrix)

set -e  # Exit on error

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
test_endpoint() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local headers="$5"

    echo -e "${BLUE}Testing:${NC} $name"

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
        echo "$body"  # Return body for variable capture
    else
        echo -e "${RED}✗ FAILED${NC} (HTTP $http_code)"
        echo "Response: $body"
        echo
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}TrainBlink Geofence + Matrix Hybrid Test${NC}"
echo -e "${YELLOW}========================================${NC}"
echo

# Test 1: Ping
echo -e "${BLUE}[Test 1] Health Check${NC}"
test_endpoint "Ping" "GET" "/ping" "" ""
echo

# Test 2: Health Status
test_endpoint "Health" "GET" "/health" "" ""
echo

# Test 3: Enter Station (User 1 - Tokyo) with P2P + Matrix
echo -e "${BLUE}[Test 3] User 1 Enters Tokyo Station (P2P + Matrix)${NC}"
user1_enter_response=$(test_endpoint \
    "User 1 Enter Station (Hybrid)" \
    "POST" \
    "/api/v1/geofence/enter" \
    '{
        "station_id": "station_tokyo_001",
        "coordinates": {
            "latitude": 35.6812,
            "longitude": 139.7671,
            "accuracy": 10.5
        },
        "client_version": "2.5.0",
        "capabilities": {
            "p2p_enabled": true,
            "matrix_enabled": true,
            "mls_supported": false
        }
    }' \
    '-H "X-User-ID: user-001" -H "X-Device-ID: iOS-123"')

# Extract session ID
user1_session=$(echo "$user1_enter_response" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}User 1 Session ID: $user1_session${NC}"
echo

# Test 4: Enter Station (User 2 - Tokyo) with P2P + Matrix
echo -e "${BLUE}[Test 4] User 2 Enters Tokyo Station (P2P + Matrix)${NC}"
user2_enter_response=$(test_endpoint \
    "User 2 Enter Station (Hybrid)" \
    "POST" \
    "/api/v1/geofence/enter" \
    '{
        "station_id": "station_tokyo_001",
        "coordinates": {
            "latitude": 35.6815,
            "longitude": 139.7670,
            "accuracy": 12.0
        },
        "client_version": "2.5.0",
        "capabilities": {
            "p2p_enabled": true,
            "matrix_enabled": true,
            "mls_supported": false
        }
    }' \
    '-H "X-User-ID: user-002" -H "X-Device-ID: Android-456"')

user2_session=$(echo "$user2_enter_response" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}User 2 Session ID: $user2_session${NC}"
echo

# Test 5: Enter Station (User 3 - Taipei) with Matrix only
echo -e "${BLUE}[Test 5] User 3 Enters Taipei Station (Matrix only)${NC}"
user3_enter_response=$(test_endpoint \
    "User 3 Enter Station (Matrix only)" \
    "POST" \
    "/api/v1/geofence/enter" \
    '{
        "station_id": "station_taipei_001",
        "coordinates": {
            "latitude": 25.0478,
            "longitude": 121.5170,
            "accuracy": 8.0
        },
        "client_version": "2.5.0",
        "capabilities": {
            "p2p_enabled": false,
            "matrix_enabled": true,
            "mls_supported": false
        }
    }' \
    '-H "X-User-ID: user-003" -H "X-Device-ID: iOS-789"')

user3_session=$(echo "$user3_enter_response" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}User 3 Session ID: $user3_session${NC}"
echo

# Test 6: Enter Station (User 4 - Tokyo) with P2P only
echo -e "${BLUE}[Test 6] User 4 Enters Tokyo Station (P2P only)${NC}"
user4_enter_response=$(test_endpoint \
    "User 4 Enter Station (P2P only)" \
    "POST" \
    "/api/v1/geofence/enter" \
    '{
        "station_id": "station_tokyo_001",
        "coordinates": {
            "latitude": 35.6810,
            "longitude": 139.7668,
            "accuracy": 15.0
        },
        "client_version": "2.5.0",
        "capabilities": {
            "p2p_enabled": true,
            "matrix_enabled": false,
            "mls_supported": false
        }
    }' \
    '-H "X-User-ID: user-004" -H "X-Device-ID: Android-999"')

user4_session=$(echo "$user4_enter_response" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}User 4 Session ID: $user4_session${NC}"
echo

# Test 7: Get Statistics
echo -e "${BLUE}[Test 7] Get Service Statistics${NC}"
test_endpoint "Get Statistics" "GET" "/api/v1/geofence/stats" "" ""
echo

# Wait a bit to simulate activity
sleep 1

# Test 8: Exit Station (User 1)
echo -e "${BLUE}[Test 8] User 1 Exits Tokyo Station${NC}"
test_endpoint \
    "User 1 Exit Station" \
    "POST" \
    "/api/v1/geofence/exit" \
    "{
        \"session_id\": \"$user1_session\",
        \"station_id\": \"station_tokyo_001\",
        \"duration_seconds\": 120,
        \"activity\": {
            \"p2p_chats_created\": 2,
            \"p2p_messages_sent\": 15,
            \"content_shared\": 3,
            \"matrix_messages_sent\": 8
        }
    }" \
    '-H "X-User-ID: user-001"'
echo

# Test 9: Exit Station (User 2)
echo -e "${BLUE}[Test 9] User 2 Exits Tokyo Station${NC}"
test_endpoint \
    "User 2 Exit Station" \
    "POST" \
    "/api/v1/geofence/exit" \
    "{
        \"session_id\": \"$user2_session\",
        \"station_id\": \"station_tokyo_001\",
        \"duration_seconds\": 90,
        \"activity\": {
            \"p2p_chats_created\": 1,
            \"p2p_messages_sent\": 12,
            \"content_shared\": 1,
            \"matrix_messages_sent\": 5
        }
    }" \
    '-H "X-User-ID: user-002"'
echo

# Test 10: Exit Station (User 3)
echo -e "${BLUE}[Test 10] User 3 Exits Taipei Station${NC}"
test_endpoint \
    "User 3 Exit Station" \
    "POST" \
    "/api/v1/geofence/exit" \
    "{
        \"session_id\": \"$user3_session\",
        \"station_id\": \"station_taipei_001\",
        \"duration_seconds\": 60,
        \"activity\": {
            \"p2p_chats_created\": 0,
            \"p2p_messages_sent\": 0,
            \"content_shared\": 0,
            \"matrix_messages_sent\": 10
        }
    }" \
    '-H "X-User-ID: user-003"'
echo

# Test 11: Exit Station (User 4)
echo -e "${BLUE}[Test 11] User 4 Exits Tokyo Station${NC}"
test_endpoint \
    "User 4 Exit Station" \
    "POST" \
    "/api/v1/geofence/exit" \
    "{
        \"session_id\": \"$user4_session\",
        \"station_id\": \"station_tokyo_001\",
        \"duration_seconds\": 45,
        \"activity\": {
            \"p2p_chats_created\": 1,
            \"p2p_messages_sent\": 5,
            \"content_shared\": 0,
            \"matrix_messages_sent\": 0
        }
    }" \
    '-H "X-User-ID: user-004"'
echo

# Test 12: Final Statistics
echo -e "${BLUE}[Test 12] Final Service Statistics${NC}"
test_endpoint "Final Statistics" "GET" "/api/v1/geofence/stats" "" ""
echo

# Summary
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}✓ Passed:${NC} $TESTS_PASSED"
echo -e "${RED}✗ Failed:${NC} $TESTS_FAILED"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
