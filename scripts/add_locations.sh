#!/bin/bash

BASE_URL="http://localhost:8080/api/v1/drivers"
OTP_CODE="123456" # assuming OTP verification bypassed or mocked as 123456

# 10 driver phone numbers (slightly different)
phones=(
  "01875113841"
  "01875113842"
  "01875113843"
  "01875113844"
  "01875113845"
  "01875113846"
  "01875113847"
  "01875113848"
  "01875113849"
  "01875113850"
)

# Corresponding locations (Dhaka area + 1 far away)
latitudes=(23.7925 23.7940 23.7461 23.8223 23.8759 23.7805 23.7610 23.7678 23.7748 24.8103)
longitudes=(90.4078 90.4043 90.3742 90.3654 90.3795 90.4264 90.3580 90.4013 90.3629 91.4125)

for i in ${!phones[@]}; do
  phone=${phones[$i]}
  lat=${latitudes[$i]}
  lng=${longitudes[$i]}

  echo "=== Registering driver ${phone} ==="
  curl -s --location "${BASE_URL}/register" \
    --header 'Content-Type: application/json' \
    --data "{\"phone\": \"${phone}\"}" > /dev/null

  echo "Requesting OTP for ${phone}"
  curl -s --location "${BASE_URL}/login/request-otp" \
    --header 'Content-Type: application/json' \
    --data "{\"phone\": \"${phone}\"}" > /dev/null

  echo "Verifying OTP for ${phone}"
  TOKEN=$(curl -s --location "${BASE_URL}/login/verify-otp" \
    --header 'Content-Type: application/json' \
    --data "{\"otp\": \"${OTP_CODE}\", \"phone\": \"${phone}\"}" \
    | jq -r '.token')

  if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed to get token for ${phone}"
    continue
  fi

  echo "Updating location for ${phone} -> lat=${lat}, lng=${lng}"
  curl -s --location "${BASE_URL}/location" \
    --header 'Content-Type: application/json' \
    --header "Authorization: Bearer ${TOKEN}" \
    --data "{\"latitude\": ${lat}, \"longitude\": ${lng}}" > /dev/null

  echo "Driver ${phone} location updated"
  echo
done

echo "All drivers registered, logged in, and location updated!"
