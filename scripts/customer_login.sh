curl --location 'http://localhost:8080/api/v1/customers/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "kaium22@gmail.com",
    "password": "123456"
}'