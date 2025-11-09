#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
OTP_CODE="123456" # assuming OTP verification bypassed or mocked as 123456

# Driver phone numbers
phones=(
  "01875114841"
)

# Optional: fixed coordinates
latitudes=(23.8103)
longitudes=(90.4125)

for i in ${!phones[@]}; do
  phone=${phones[$i]}
  lat=${latitudes[$i]}
  lng=${longitudes[$i]}

  echo "=== Registering driver ${phone} ==="
  curl -s --location "${BASE_URL}/drivers/register" \
    --header "Content-Type: application/json" \
    --data "{\"phone\": \"${phone}\"}" > /dev/null

  echo "Requesting OTP for ${phone}"
  curl -s --location "${BASE_URL}/drivers/login/request-otp" \
    --header "Content-Type: application/json" \
    --data "{\"phone\": \"${phone}\"}" > /dev/null

  echo "Verifying OTP for ${phone}"
  TOKEN=$(curl -s --location "${BASE_URL}/drivers/login/verify-otp" \
    --header "Content-Type: application/json" \
    --data "{\"otp\": \"${OTP_CODE}\", \"phone\": \"${phone}\"}" \
    | jq -r '.token')

  if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed to get token for ${phone}"
    continue
  fi

  echo "Token received for ${phone}"

  # Find nearby rides (correct endpoint)
  echo "Finding nearest rides for driver ${phone}..."
  curl --location "${BASE_URL}/rides/nearby" \
    --header "Authorization: Bearer $TOKEN" \
    --header "Content-Type: application/json" \
    --data "{
      \"lat\": ${lat},
      \"lng\": ${lng},
      \"radius\": 5000,
      \"limit\": 3
    }" \
    --silent | jq || echo "No rides found"

  echo "Driver ${phone} location checked"
  echo
  sleep 1
done

echo "All drivers registered, logged in, and nearby rides fetched!"
