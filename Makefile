run:
	go run main.go

req:
	curl -H "Content-Type: application/json"  -d '{"method":"Arith.Divide","params":[{"A": 10, "B":2}], "id": 0}' http://localhost:8080 -v
