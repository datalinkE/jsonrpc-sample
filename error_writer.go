package rpcserver

import (
	"github.com/gorilla/rpc/v2"
	jsonrpc "github.com/gorilla/rpc/v2/json2"
	"net/http"
)

type ProtocolErrorWriter interface {
	ProtocolError(w http.ResponseWriter, codecReq rpc.CodecRequest, status int, msg string)
}

type JSONRPC2ErrorWriter struct {
}

func (ew *JSONRPC2ErrorWriter) ProtocolError(w http.ResponseWriter, codecReq rpc.CodecRequest, status int, msg string) {
	codecReq.WriteError(w, status, &jsonrpc.Error{
		Code:    jsonrpc.ErrorCode(status),
		Message: msg,
	})
}
