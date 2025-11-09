curl --location 'http://localhost:8080/api/v1/drivers/login/verify-otp' \
--header 'Content-Type: application/json' \
--data '{
    "otp": "123456",
    "phone": "01875113947"
}'