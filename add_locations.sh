# Driver 1 - Gulshan, Dhaka

token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJyb2xlIjoiZHJpdmVyIiwiZXhwIjoxNzY2MTAxMzgxLCJuYmYiOjE3NjI1MDEzODEsImlhdCI6MTc2MjUwMTM4MX0.D1MtNtjuUQS_hklxT921nO3O6wn5a6GACNcMdsWdocA"
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.7925, "longitude": 90.4078}'

# Driver 2 - Banani, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.7940, "longitude": 90.4043}'

# Driver 3 - Dhanmondi, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.7461, "longitude": 90.3742}'

# Driver 4 - Mirpur, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.8223, "longitude": 90.3654}'

# Driver 5 - Uttara, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.8759, "longitude": 90.3795}'

# Driver 6 - Badda, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.7805, "longitude": 90.4264}'

# Driver 7 - Mohammadpur, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.7610, "longitude": 90.3580}'

# Driver 8 - Tejgaon, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.7678, "longitude": 90.4013}'

# Driver 9 - Shyamoli, Dhaka
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 23.7748, "longitude": 90.3629}'

# Driver 10 - Sylhet (for far-away test)
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $token" \
--data '{"latitude": 24.8103, "longitude": 91.4125}'
