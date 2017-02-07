package methodcodec

import (
	"fmt"
	"github.com/gorilla/rpc/v2"
	jsonrpc "github.com/gorilla/rpc/v2/json2"
	"net/http"
	"strings"
)

type (
	NestedCodecRequest struct {
		super rpc.CodecRequest
		err   error
	}

	MethodCodec struct {
		super *jsonrpc.Codec
	}
)

func NewNestedCodecRequest(super rpc.CodecRequest, err error) *NestedCodecRequest {
	return &NestedCodecRequest{super: super, err: err}
}

func (c *NestedCodecRequest) Method() (string, error) {
	if c.err != nil {
		return "", c.err
	}
	return c.super.Method()
}

func (c *NestedCodecRequest) ReadRequest(args interface{}) error {
	if c.err != nil {
		return c.err
	}
	return c.super.ReadRequest(args)
}

func (c *NestedCodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}) {
	c.super.WriteResponse(w, reply)
}

func (c *NestedCodecRequest) WriteError(w http.ResponseWriter, status int, err error) {
	c.super.WriteError(w, status, err)
}

func NewMethodCodec() *MethodCodec {
	return &MethodCodec{super: jsonrpc.NewCodec()}
}

func (c *MethodCodec) NewRequest(r *http.Request) rpc.CodecRequest {
	result := c.super.NewRequest(r)
	path := r.URL.Path
	method, err := result.Method()
	if err != nil {
		return result
	}

	if !PathHasMethod(path, method) {
		err = fmt.Errorf("URL.Path '%v' does not end with method '%v'", path, method)
	}

	return &NestedCodecRequest{
		super: result,
		err:   err,
	}
}

func PathHasMethod(path string, method string) bool {
	pathWords := strings.Split(path, "/")
	if method == "" || len(pathWords) < 1 {
		return false
	}

	return pathWords[len(pathWords)-1] == method
}
