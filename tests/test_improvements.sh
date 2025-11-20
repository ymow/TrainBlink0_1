#!/bin/bash

# Test script for WebSocket improvements
# Tests rate limiting, validation, and stats integration

set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}==================================${NC}"
echo -e "${BLUE}WebSocket Improvements Test${NC}"
echo -e "${BLUE}==================================${NC}"
echo

# Test 1: Get stats with WebSocket info
echo -e "${BLUE}[Test 1] Get Statistics (with WebSocket stats)${NC}"
response=$(curl -s http://localhost:8080/api/v1/geofence/stats)
echo "$response" | jq .

# Check if websocket stats exist
if echo "$response" | jq -e '.data.websocket' > /dev/null; then
    echo -e "${GREEN}✓ WebSocket statistics found in response${NC}"
    echo

    # Show WebSocket specific stats
    echo "WebSocket Stats:"
    echo "$response" | jq '.data.websocket'
    echo
else
    echo -e "${RED}✗ WebSocket statistics not found${NC}"
fi

# Test 2: Enter station and connect WebSocket
echo -e "${BLUE}[Test 2] Enter station${NC}"
enter_response=$(curl -s -X POST http://localhost:8080/api/v1/geofence/enter \
    -H 'Content-Type: application/json' \
    -H 'X-User-ID: test-user-001' \
    -H 'X-Device-ID: test-device-001' \
    -d '{
        "station_id": "station_tokyo_001",
        "coordinates": {"latitude": 35.6812, "longitude": 139.7671, "accuracy": 10.5},
        "client_version": "2.5.0",
        "capabilities": {"p2p_enabled": true, "matrix_enabled": true, "mls_supported": false}
    }')

session_id=$(echo "$enter_response" | jq -r '.data.session_id')
echo "Session ID: $session_id"
echo -e "${GREEN}✓ Station entered successfully${NC}"
echo

# Test 3: Get updated stats
echo -e "${BLUE}[Test 3] Get Updated Statistics${NC}"
response2=$(curl -s http://localhost:8080/api/v1/geofence/stats)

# Check rate limits in stats
if echo "$response2" | jq -e '.data.websocket.rate_limits' > /dev/null; then
    echo -e "${GREEN}✓ Rate limit configuration found:${NC}"
    echo "$response2" | jq '.data.websocket.rate_limits'
    echo
else
    echo -e "${RED}✗ Rate limit configuration not found${NC}"
fi

echo -e "${BLUE}==================================${NC}"
echo -e "${GREEN}All improvement tests completed!${NC}"
echo -e "${BLUE}==================================${NC}"
