#!/bin/bash

# Demo script for Pub/Sub System
# Make sure the server is running on localhost:8080

echo "=== Pub/Sub System Demo ==="
echo "Server should be running on localhost:8080"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to make HTTP requests and show results
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${YELLOW}${description}${NC}"
    echo "curl -X ${method} http://localhost:8080${endpoint}"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
        response=$(curl -s -X ${method} -H "Content-Type: application/json" -d "$data" http://localhost:8080${endpoint})
    else
        response=$(curl -s -X ${method} http://localhost:8080${endpoint})
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Response:${NC} $response"
    else
        echo -e "${RED}Error: Request failed${NC}"
    fi
    echo ""
}

# Check if server is running
echo "Checking if server is running..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}Server is running!${NC}"
    echo ""
else
    echo -e "${RED}Server is not running. Please start the server first.${NC}"
    echo "Run: ./pub-sub"
    exit 1
fi

# Demo sequence
echo "=== Starting Demo ==="
echo ""

# 1. Check health
make_request "GET" "/health" "" "1. Checking system health"

# 2. Create a topic
make_request "POST" "/topics" '{"name": "orders"}' "2. Creating 'orders' topic"

# 3. Create another topic
make_request "POST" "/topics" '{"name": "notifications"}' "3. Creating 'notifications' topic"

# 4. List all topics
make_request "GET" "/topics" "" "4. Listing all topics"

# 5. Publish a message to orders topic
make_request "POST" "/publish" '{"topic": "orders", "message": {"id": "order-123", "payload": {"order_id": "ORD-123", "amount": 99.50, "currency": "USD"}}}' "5. Publishing message to 'orders' topic"

# 6. Publish another message
make_request "POST" "/publish" '{"topic": "orders", "message": {"id": "order-124", "payload": {"order_id": "ORD-124", "amount": 150.00, "currency": "EUR"}}}' "6. Publishing another message to 'orders' topic"

# 7. Publish to notifications topic
make_request "POST" "/publish" '{"topic": "notifications", "message": {"id": "notif-1", "payload": {"type": "info", "message": "System maintenance scheduled"}}}' "7. Publishing message to 'notifications' topic"

# 8. Get stats
make_request "GET" "/stats" "" "8. Getting system statistics"

# 9. Get health again
make_request "GET" "/health" "" "9. Checking health again"

# 10. Delete a topic
make_request "DELETE" "/topics?name=notifications" "" "10. Deleting 'notifications' topic"

# 11. List topics again
make_request "GET" "/topics" "" "11. Listing topics after deletion"

echo "=== Demo Complete ==="
echo ""
echo "To test WebSocket functionality:"
echo "1. Open test-client.html in a browser"
echo "2. Connect to ws://localhost:8080/ws"
echo "3. Subscribe to topics and publish messages"
echo ""
echo "For more examples, see the README.md file"
