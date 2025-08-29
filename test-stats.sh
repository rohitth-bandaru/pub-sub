#!/bin/bash

echo "🧪 Testing Pub/Sub Stats API"
echo "=============================="

# Base URL
BASE_URL="http://localhost:8080"

echo ""
echo "📊 1. Initial Stats (should be empty)"
echo "----------------------------------------"
curl -s "$BASE_URL/stats" | jq '.'

echo ""
echo "📝 2. Creating Topics"
echo "----------------------"

# Create test topics
echo "Creating 'orders' topic..."
curl -s -X POST "$BASE_URL/topics" \
  -H "Content-Type: application/json" \
  -d '{"name": "orders"}' | jq '.'

echo "Creating 'users' topic..."
curl -s -X POST "$BASE_URL/topics" \
  -H "Content-Type: application/json" \
  -d '{"name": "users"}' | jq '.'

echo "Creating 'notifications' topic..."
curl -s -X POST "$BASE_URL/topics" \
  -H "Content-Type: application/json" \
  -d '{"name": "notifications"}' | jq '.'

echo ""
echo "📊 3. Stats After Creating Topics"
echo "----------------------------------"
curl -s "$BASE_URL/stats" | jq '.'

echo ""
echo "📤 4. Publishing Messages"
echo "--------------------------"

# Publish messages to topics
echo "Publishing message to 'orders' topic..."
curl -s -X POST "$BASE_URL/publish" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "orders",
    "message": {
      "id": "order-1",
      "payload": {"order_id": "ORD-123", "amount": 99.99}
    }
  }' | jq '.'

echo "Publishing message to 'users' topic..."
curl -s -X POST "$BASE_URL/publish" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "users",
    "message": {
      "id": "user-1",
      "payload": {"user_id": "USR-001", "name": "John Doe"}
    }
  }' | jq '.'

echo "Publishing another message to 'orders' topic..."
curl -s -X POST "$BASE_URL/publish" \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "orders",
    "message": {
      "id": "order-2",
      "payload": {"order_id": "ORD-124", "amount": 149.99}
    }
  }' | jq '.'

echo ""
echo "📊 5. Final Stats After Publishing Messages"
echo "--------------------------------------------"
curl -s "$BASE_URL/stats" | jq '.'

echo ""
echo "📋 6. List All Topics"
echo "----------------------"
curl -s "$BASE_URL/topics" | jq '.'

echo ""
echo "🏥 7. System Health"
echo "-------------------"
curl -s "$BASE_URL/health" | jq '.'

echo ""
echo "🎉 Stats API Test Completed!"
echo ""
echo "Expected Results:"
echo "- Total Topics: 3"
echo "- Total Messages: 3"
echo "- Total Subscribers: 0 (no WebSocket clients connected)"
echo "- Active Connections: 0 (no WebSocket clients connected)"
