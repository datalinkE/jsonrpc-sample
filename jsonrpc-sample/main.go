package main

import (
	"errors"
	"github.com/datalinkE/rpcserver"
	"github.com/datalinkE/rpcserver/jsonrpc2"
	"gopkg.in/gin-gonic/gin.v1"
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

	anotherServer, err := rpcserver.NewServer(arith)
	if err != nil {
		log.Fatal(err)
	}
	anotherServer.RegisterCodec(jsonrpc2.NewCodec(), "application/json")

	router := gin.Default()
	router.POST("/jsonrpc/v2/:method", gin.WrapH(anotherServer))

	log.Fatal(router.Run())
}
