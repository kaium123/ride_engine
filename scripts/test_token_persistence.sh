#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "=========================================="
echo "Testing JWT Token Persistence"
echo "=========================================="

# Step 1: Login and get token
echo ""
echo "1️⃣  Logging in customer..."
LOGIN_RESP=$(curl -s --location "${BASE_URL}/customers/login" \
  --header "Content-Type: application/json" \
  --data '{
    "email": "kaium@gmail.com",
    "password": "123456"
  }')

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.token // .access_token // .data.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Failed to get token. Response:"
  echo "$LOGIN_RESP"
  exit 1
fi

echo "✅ Token received: ${TOKEN:0:50}..."

# Step 2: Create a ride (this should work)
echo ""
echo "2️⃣  Creating a ride with the token..."
RIDE_RESP=$(curl -s --location "${BASE_URL}/rides/" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $TOKEN" \
  --data '{
    "pickup_lat": 23.7801,
    "pickup_lng": 90.4050,
    "dropoff_lat": 23.8001,
    "dropoff_lng": 90.4250"
  }')

echo "Response:"
echo "$RIDE_RESP" | jq

RIDE_ID=$(echo "$RIDE_RESP" | jq -r '.id // empty')

if [ -z "$RIDE_ID" ]; then
  echo "❌ Failed to create ride"
  exit 1
fi

echo "✅ Ride created with ID: $RIDE_ID"

# Step 3: Immediately check ride status with same token
echo ""
echo "3️⃣  Checking ride status immediately with same token..."
echo "Using token: ${TOKEN:0:50}..."

STATUS_RESP=$(curl -s --location "${BASE_URL}/rides/status?ride_id=${RIDE_ID}" \
  --header "Authorization: Bearer $TOKEN" \
  --header "Content-Type: application/json")

echo "Response:"
echo "$STATUS_RESP" | jq

# Check if successful
if echo "$STATUS_RESP" | jq -e '.error' > /dev/null 2>&1; then
  echo "❌ Failed! Got error: $(echo "$STATUS_RESP" | jq -r '.error')"
  exit 1
else
  echo "✅ Success! Token persisted correctly"
fi

echo ""
echo "=========================================="
echo "Test completed!"
echo "=========================================="
