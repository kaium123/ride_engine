curl --location 'http://localhost:8080/api/v1/customers/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "kaium22@gmail.com",
    "phone": "01875113821",
    "name": "kaium",
    "password": "123456"
}'