run:
	go run main.go

req:
	curl -H "Content-Type: application/json"  -d '{"jsonrpc": "2.0", "method":"Divide","params":[{"A": 10, "B":2}], "id": 1}' http://localhost:8080/jsonrpc/v2/Divide -v

req-wrong:
	curl -H "Content-Type: application/json"  -d '{"jsonrpc": "2.0", "method":"Wrong","params":[{"A": 10, "B":2}], "id": 1}' http://localhost:8080/jsonrpc/v2/Wrong -v

req-incorrect:
	curl -H "Content-Type: application/json"  -d '{"jsonrpc": "2.0", "method":"Wrong","params":[{"A": 10, "B":2}], "id": 1}' http://localhost:8080/jsonrpc/v2/Divide -v

