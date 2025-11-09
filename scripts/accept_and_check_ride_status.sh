#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
OTP_CODE="123456" # assuming OTP verification bypassed or mocked

# === Customer Setup (done once) ===
# Only register customer if they don't exist yet
 echo "Registering customer..."
 curl -s --location "${BASE_URL}/customers/register" \
   --header "Content-Type: application/json" \
   --data '{
     "email": "kaium@gmail.com",
     "phone": "01875113843",
     "name": "kaium",
     "password": "123456"
   }'

echo "Logging in customer..."
CUSTOMER_LOGIN_RESP=$(curl -s --location "${BASE_URL}/customers/login" \
  --header "Content-Type: application/json" \
  --data '{
    "email": "kaium@gmail.com",
    "password": "123456"
  }')

CUSTOMER_TOKEN=$(echo "$CUSTOMER_LOGIN_RESP" | jq -r '.token // .access_token // .data.token')

if [ "$CUSTOMER_TOKEN" == "null" ] || [ -z "$CUSTOMER_TOKEN" ]; then
  echo "Failed to get customer token. Check login response:"
  echo "$CUSTOMER_LOGIN_RESP"
  exit 1
fi

echo "Customer token received!"

# === Customer requests a ride ===
echo "Requesting a new ride..."
RIDE_CREATE_RESP=$(curl -s --location "${BASE_URL}/rides/" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $CUSTOMER_TOKEN" \
  --data '{
    "pickup_lat": 23.7801,
    "pickup_lng": 90.4050,
    "dropoff_lat": 23.8001,
    "dropoff_lng": 90.4250
  }')

echo "Ride creation response:"
echo "$RIDE_CREATE_RESP" | jq

echo "$CUSTOMER_TOKEN"

# === Driver setup ===
phones=("01875113841")
latitudes=(23.8103)
longitudes=(90.4125)

for i in ${!phones[@]}; do
  phone=${phones[$i]}
  lat=${latitudes[$i]}
  lng=${longitudes[$i]}

  echo ""
  echo "==============================="
  echo "Registering driver ${phone}"
  echo "==============================="

  curl -s --location "${BASE_URL}/drivers/register" \
    --header "Content-Type: application/json" \
    --data "{\"phone\": \"${phone}\"}" > /dev/null

  echo "📲 Requesting OTP for ${phone}..."
  curl -s --location "${BASE_URL}/drivers/login/request-otp" \
    --header "Content-Type: application/json" \
    --data "{\"phone\": \"${phone}\"}" > /dev/null

  echo "Verifying OTP for ${phone}..."
  TOKEN=$(curl -s --location "${BASE_URL}/drivers/login/verify-otp" \
    --header "Content-Type: application/json" \
    --data "{\"otp\": \"${OTP_CODE}\", \"phone\": \"${phone}\"}" \
    | jq -r '.token')

  if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed to get token for ${phone}"
    continue
  fi

  echo "Token received for driver ${phone}"

  # === Step 1: Find nearby rides ===
  echo "Finding nearby rides for driver ${phone}..."
  RIDES_JSON=$(curl -s --location "${BASE_URL}/rides/nearby" \
    --header "Authorization: Bearer $TOKEN" \
    --header "Content-Type: application/json" \
    --data "{
      \"lat\": ${lat},
      \"lng\": ${lng},
      \"radius\": 5000,
      \"limit\": 3
    }")

  echo "$RIDES_JSON" | jq

  # === Step 2: Pick first available ride ===
  ride_id=$(echo "$RIDES_JSON" | jq -r '.[0].id // empty')

  if [ -z "$ride_id" ]; then
    echo "No rides available near driver ${phone}"
    continue
  fi

  echo "Selected ride ID: $ride_id"
  echo "Accepting ride $ride_id for driver ${phone}..."

  response=$(curl -s --location --request POST "${BASE_URL}/rides/accept?ride_id=${ride_id}" \
    --header "Authorization: Bearer $TOKEN" \
    --data '')

  echo "📦 Accept response:"
  echo "$response" | jq

  echo "$CUSTOMER_TOKEN"
  # === Step 3: Check ride status ===
  echo "🔎 Checking ride status for ID $ride_id..."
  status_response=$(curl -s --location "${BASE_URL}/rides/status?ride_id=${ride_id}" \
    --header "Authorization: Bearer $CUSTOMER_TOKEN" \
    --header "Content-Type: application/json")

  echo "Ride status response:"
  echo "$status_response" | jq

done

echo ""
echo "All drivers registered, logged in, and rides accepted successfully!"
