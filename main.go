package main

import (
	"errors"
	"github.com/gorilla/rpc"
	jsonrpc "github.com/gorilla/rpc/json"
	"log"
	"net/http"
)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(r *http.Request, args *Args, reply *int) error {
	log.Print("multiply")
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(r *http.Request, args *Args, quo *Quotient) error {
	log.Print("divide")
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

func main() {
	log.Print("main")
	arith := new(Arith)
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	rpcServer.RegisterService(arith, "Arith")
	http.Handle("/", rpcServer)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
