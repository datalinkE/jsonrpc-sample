# jsonrpc-sample

jsonrpc-sample demonstrates how to serve [Golang](http://golang.org) RPC methods over HTTP using the [JSON-RPC](http://golang.org/gorilla/rpc/json). 

## Usage

### Install

Start the jsonrpc sample:

```
go run main.go
or
make run
```

### Make RPC calls using curl

```
curl -d '{"method":"Arith.Divide","params":[{"A": 10, "B":2}], "id": 0}' http://localhost:8080
or
make req
```

```
{"id":0,"result":{"Quo":5,"Rem":0},"error":null}
```
