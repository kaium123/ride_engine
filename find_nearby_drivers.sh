#!/bin/bash

API_BASE="http://localhost:8080/api/v1"

# 1️⃣ Register customer
echo "👤 Registering customer..."
curl -s --location "$API_BASE/customers/register" \
  --header 'Content-Type: application/json' \
  --data '
  {
      "email": "kaium@gmail.com",
      "phone": "01875113843",
      "name": "kaium",
      "password": "123456"
  }
  ' > /dev/null

# 2️⃣ Login customer
echo "🔑 Logging in customer..."
CUSTOMER_LOGIN_RESP=$(curl -s --location "$API_BASE/customers/login" \
  --header 'Content-Type: application/json' \
  --data '{
              "email": "kaium@gmail.com",
              "password": "123456"
          }')

echo "🧩 Raw login response: $CUSTOMER_LOGIN_RESP"

# Try to extract token — adjust based on your API’s actual response
CUSTOMER_TOKEN=$(echo "$CUSTOMER_LOGIN_RESP" | jq -r '.token // .access_token // .data.token')

if [ "$CUSTOMER_TOKEN" == "null" ] || [ -z "$CUSTOMER_TOKEN" ]; then
  echo "❌ Failed to get customer token. Please check your login response above."
  exit 1
fi

echo "✅ Got token: ${CUSTOMER_TOKEN:0:30}..."

# 3️⃣ Find nearby drivers
echo "🚗 Finding nearest drivers for customer..."
curl --location "$API_BASE/rides/nearby" \
  --header "Authorization: Bearer $CUSTOMER_TOKEN" \
  --header "Content-Type: application/json" \
  --data '{
    "latitude": 23.8103,
    "longitude": 90.4125,
    "radius": 5000,
    "limit": 3
  }' \
  --silent | jq


