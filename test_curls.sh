curl -X POST "http://localhost:8888/v1/user/signup" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "email": "test@example.com", "password": "Pass@12345"}'

curl -X POST "http://localhost:8888/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com", "password":"Pass@12345"}' 


curl -X POST \
  http://localhost:8888/v1/expenses \
  -H 'Content-Type: application/json' \
  -d '{
    "description": "test expense",
    "amount": 123.45,
    "split": {"user1":"25%","user2":"75%"},
    "payee": {"name":"John Doe","email":"johndoe@example.com"}
  }'

curl -X POST \
  http://localhost:8888/api/groups \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "test group",
    "description": "This is a test group",
    "admin_id": "userId"
  }'

curl -X POST \
  http://localhost:8888/api/groups/{groupId}/invite \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "userId"
  }'
