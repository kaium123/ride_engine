#!/bin/bash

BASE_URL="http://localhost:8080/api/v1/customers"
RIDE_URL="http://localhost:8080/api/v1/rides/"
PASSWORD="123456"

# === 📧 Customer Emails & Phones (10 customers) ===
emails=(
  "kaium0@mail.com" "kaium1@mail.com" "kaium2@mail.com" "kaium3@mail.com" "kaium4@mail.com"
  "kaium5@mail.com" "kaium6@mail.com" "kaium7@mail.com" "kaium8@mail.com" "kaium9@mail.com"
)
phones=(
  "01875113840" "01875113841" "01875113842" "01875113843" "01875113844"
  "01875113845" "01875113846" "01875113847" "01875113848" "01875113849"
)

# === 📍 Pickup & Dropoff Coordinates (10 rides) ===
pickup_lats=(23.7801 23.7815 23.7830 23.7842 23.7855 23.7868 23.7881 23.7893 23.7906 23.7919)
pickup_lngs=(90.4050 90.4062 90.4074 90.4086 90.4098 90.4110 90.4122 90.4134 90.4146 90.4158)
dropoff_lats=(23.8001 23.8015 23.8030 23.8042 23.8055 23.8068 23.8081 23.8093 23.8106 23.8119)
dropoff_lngs=(90.4250 90.4262 90.4274 90.4286 90.4298 90.4310 90.4322 90.4334 90.4346 90.4358)

echo "👥 Creating 10 users and rides..."
echo

for i in "${!emails[@]}"; do
  email=${emails[$i]}
  phone=${phones[$i]}
  name="Kaium${i}"

  echo "=== 🧍 Processing ${name} (${email}) ==="

  # Try to register
  register_response=$(curl -s -w "%{http_code}" -o /tmp/reg_resp.json \
    --location "${BASE_URL}/register" \
    --header 'Content-Type: application/json' \
    --data "{\"name\": \"${name}\", \"email\": \"${email}\", \"phone\": \"${phone}\", \"password\": \"${PASSWORD}\"}")

  if [ "$register_response" -eq 200 ]; then
    echo "✅ Registered ${email} successfully."
  elif [ "$register_response" -eq 409 ]; then
    echo "ℹ️ ${email} already exists — skipping registration."
  else
    echo "⚠️ Registration failed for ${email} (HTTP ${register_response})."
  fi

  # Login
  echo "🔑 Logging in ${email}..."
  login_response=$(curl -s --location "${BASE_URL}/login" \
    --header 'Content-Type: application/json' \
    --data "{\"email\": \"${email}\", \"password\": \"${PASSWORD}\"}")

  token=$(echo "$login_response" | jq -r '.token // .data.token')

  if [ "$token" == "null" ] || [ -z "$token" ]; then
    echo "❌ Failed to login for ${email}. Response: $login_response"
    echo
    continue
  fi
  echo "✅ Logged in ${email} successfully."

  # Create ride
  pickup_lat=${pickup_lats[$i]}
  pickup_lng=${pickup_lngs[$i]}
  dropoff_lat=${dropoff_lats[$i]}
  dropoff_lng=${dropoff_lngs[$i]}

  echo "🚗 Creating ride: (${pickup_lat}, ${pickup_lng}) → (${dropoff_lat}, ${dropoff_lng})"
  ride_response=$(curl -s --location "${RIDE_URL}" \
    --header "Authorization: Bearer ${token}" \
    --header 'Content-Type: application/json' \
    --data "{
      \"pickup_lat\": ${pickup_lat},
      \"pickup_lng\": ${pickup_lng},
      \"dropoff_lat\": ${dropoff_lat},
      \"dropoff_lng\": ${dropoff_lng}
    }")

  echo "📦 Response: $(echo "$ride_response" | jq -r '.message // .error // .data')"
  echo
done

echo "✅ All 10 users processed and rides created successfully!"
