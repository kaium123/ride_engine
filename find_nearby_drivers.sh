
token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJyb2xlIjoiY3VzdG9tZXIiLCJleHAiOjE3NjYxMDE0NDcsIm5iZiI6MTc2MjUwMTQ0NywiaWF0IjoxNzYyNTAxNDQ3fQ.FOueiSCOQgGvrfBlOJGiklxtG7-c9HKTxrC8tlmvlpE"

curl --location 'http://localhost:8080/api/v1/rides/nearby?lat=23.8103&lng=90.4125&radius=5000&limit=3' \
--header "Authorization: Bearer $token"
