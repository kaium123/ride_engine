#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
EMAIL="kaium11@gmail.com"
PHONE="01875113811"
NAME="kaium"
PASSWORD="123456"

echo "🚀 Registering customer..."
REGISTER_RESPONSE=$(curl --silent --location "$BASE_URL/customers/register" \
  --header 'Content-Type: application/json' \
  --data-raw "{
    \"email\": \"$EMAIL\",
    \"phone\": \"$PHONE\",
    \"name\": \"$NAME\",
    \"password\": \"$PASSWORD\"
  }")

echo "📝 Register response: $REGISTER_RESPONSE"

echo ""
echo "🔐 Logging in customer..."
LOGIN_RESPONSE=$(curl --silent --location "$BASE_URL/customers/login" \
  --header 'Content-Type: application/json' \
  --data-raw "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "🧾 Login response: $LOGIN_RESPONSE"

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Failed to extract token. Exiting."
  exit 1
fi

echo "✅ Token extracted successfully."

echo ""
echo "🚗 Creating ride request..."
RIDE_RESPONSE=$(curl --silent --location "$BASE_URL/rides/" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $TOKEN" \
  --data '{
    "pickup_lat": 23.23,
    "pickup_lng": 90.24,
    "dropoff_lat": 25.23,
    "dropoff_lng": 95.24
  }')

echo "🧾 Ride response: $RIDE_RESPONSE"

echo ""
echo "🎉 Done!"
