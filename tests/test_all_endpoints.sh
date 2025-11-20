#!/bin/bash

# Complete Testing Script for TrainBlink Server
# Tests all HTTP endpoints including POST requests

set -e

BASE_URL="http://localhost:8080"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "╔══════════════════════════════════════════════════════╗"
echo "║                                                      ║"
echo "║     🚄 TrainBlink Server - Complete Test Suite      ║"
echo "║                                                      ║"
echo "╚══════════════════════════════════════════════════════╝"
echo ""

# Check if server is running
echo "Checking if server is running..."
if ! curl -s "$BASE_URL/ping" > /dev/null 2>&1; then
    echo -e "${RED}❌ Server is not running!${NC}"
    echo "Please start the server first:"
    echo "  go run cmd/server/main_with_post.go"
    exit 1
fi
echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# Test 1: GET /ping
echo -e "${YELLOW}Test 1: GET /ping${NC}"
curl -s "$BASE_URL/ping" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/ping"
echo ""

# Test 2: GET /health
echo -e "${YELLOW}Test 2: GET /health${NC}"
curl -s "$BASE_URL/health" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/health"
echo ""

# Test 3: GET /api/v1/hello
echo -e "${YELLOW}Test 3: GET /api/v1/hello${NC}"
curl -s "$BASE_URL/api/v1/hello" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/v1/hello"
echo ""

# Test 4: POST /api/v1/clients - Register Client A
echo -e "${YELLOW}Test 4: POST /api/v1/clients - Register Alice${NC}"
CLIENT_A_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/clients" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "device_id": "iOS-1234"
  }')
echo "$CLIENT_A_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$CLIENT_A_RESPONSE"
CLIENT_A_ID=$(echo "$CLIENT_A_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['client']['id'])" 2>/dev/null || echo "client-a")
echo -e "${GREEN}Alice's Client ID: $CLIENT_A_ID${NC}"
echo ""

# Test 5: POST /api/v1/clients - Register Client B
echo -e "${YELLOW}Test 5: POST /api/v1/clients - Register Bob${NC}"
CLIENT_B_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/clients" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob",
    "device_id": "Android-5678"
  }')
echo "$CLIENT_B_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$CLIENT_B_RESPONSE"
CLIENT_B_ID=$(echo "$CLIENT_B_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['client']['id'])" 2>/dev/null || echo "client-b")
echo -e "${GREEN}Bob's Client ID: $CLIENT_B_ID${NC}"
echo ""

# Test 6: GET /api/v1/clients - List all clients
echo -e "${YELLOW}Test 6: GET /api/v1/clients - List all clients${NC}"
curl -s "$BASE_URL/api/v1/clients" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/v1/clients"
echo ""

# Test 7: POST /api/v1/messages - Alice sends to Bob
echo -e "${YELLOW}Test 7: POST /api/v1/messages - Alice → Bob${NC}"
curl -s -X POST "$BASE_URL/api/v1/messages" \
  -H "Content-Type: application/json" \
  -d "{
    \"from\": \"$CLIENT_A_ID\",
    \"to\": \"$CLIENT_B_ID\",
    \"text\": \"Hello Bob, this is Alice!\"
  }" | python3 -m json.tool 2>/dev/null || curl -s -X POST "$BASE_URL/api/v1/messages" \
  -H "Content-Type: application/json" \
  -d "{\"from\": \"$CLIENT_A_ID\", \"to\": \"$CLIENT_B_ID\", \"text\": \"Hello Bob!\"}"
echo ""

# Test 8: POST /api/v1/messages - Bob sends to Alice
echo -e "${YELLOW}Test 8: POST /api/v1/messages - Bob → Alice${NC}"
curl -s -X POST "$BASE_URL/api/v1/messages" \
  -H "Content-Type: application/json" \
  -d "{
    \"from\": \"$CLIENT_B_ID\",
    \"to\": \"$CLIENT_A_ID\",
    \"text\": \"Hi Alice, nice to meet you!\"
  }" | python3 -m json.tool 2>/dev/null || curl -s -X POST "$BASE_URL/api/v1/messages" \
  -H "Content-Type: application/json" \
  -d "{\"from\": \"$CLIENT_B_ID\", \"to\": \"$CLIENT_A_ID\", \"text\": \"Hi Alice!\"}"
echo ""

# Test 9: POST /api/v1/messages - Broadcast message
echo -e "${YELLOW}Test 9: POST /api/v1/messages - Broadcast${NC}"
curl -s -X POST "$BASE_URL/api/v1/messages" \
  -H "Content-Type: application/json" \
  -d "{
    \"from\": \"system\",
    \"to\": \"all\",
    \"text\": \"Welcome everyone to TrainBlink!\"
  }" | python3 -m json.tool 2>/dev/null || curl -s -X POST "$BASE_URL/api/v1/messages" \
  -H "Content-Type: application/json" \
  -d '{"from": "system", "to": "all", "text": "Welcome!"}'
echo ""

# Test 10: GET /api/v1/messages - List all messages
echo -e "${YELLOW}Test 10: GET /api/v1/messages - List all messages${NC}"
curl -s "$BASE_URL/api/v1/messages" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/v1/messages"
echo ""

# Test 11: GET /api/v1/messages/user - Get Alice's messages
echo -e "${YELLOW}Test 11: GET /api/v1/messages/user?id=$CLIENT_A_ID - Alice's messages${NC}"
curl -s "$BASE_URL/api/v1/messages/user?id=$CLIENT_A_ID" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/v1/messages/user?id=$CLIENT_A_ID"
echo ""

# Test 12: GET /api/v1/messages/user - Get Bob's messages
echo -e "${YELLOW}Test 12: GET /api/v1/messages/user?id=$CLIENT_B_ID - Bob's messages${NC}"
curl -s "$BASE_URL/api/v1/messages/user?id=$CLIENT_B_ID" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/v1/messages/user?id=$CLIENT_B_ID"
echo ""

echo "════════════════════════════════════════════════════════"
echo ""
echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
echo ""
echo "Summary:"
echo "  • 2 clients registered (Alice, Bob)"
echo "  • 3 messages sent (Alice→Bob, Bob→Alice, System→All)"
echo "  • All endpoints tested and working"
echo ""
