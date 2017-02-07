run:
	go run main.go

req:
	curl -H "Content-Type: application/json"  -d '{"jsonrpc": "2.0", "method":"Arith.Divide","params":[{"A": 10, "B":2}], "id": 1}' http://localhost:8080/jsonrpc/v1/Arith.Divide -v
