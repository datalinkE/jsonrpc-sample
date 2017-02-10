# jsonrpc-sample

jsonrpc-sample demonstrates how to serve [Golang](http://golang.org) RPC methods over HTTP using the [JSON-RPC](http://json-rpc.org/wiki/specification). 

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
curl -H "Content-Type: application/json"  -d '{"jsonrpc": "2.0", "method":"Arith.Divide","params":[{"A": 10, "B":2}], "id": 1}' http://localhost:8080/jsonrpc/v2/Arith.Divide -v
or
make req
```

```
{"jsonrpc":"2.0","result":{"Quo":5,"Rem":0},"id":1}
```
