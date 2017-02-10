package rpcserver

import (
	"net/http"
)

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// Codec creates a CodecRequest to process each request.
type Codec interface {
	NewRequest(*http.Request) CodecRequest
}

// CodecRequest decodes a request and encodes a response using a specific
// serialization scheme.
type CodecRequest interface {
	// Shows if request was valid or incorrect. Check this after CodecRequest was created.
	Error() error
	// Reads the request and returns the RPC method name.
	Method() (string, error)
	// Reads the request filling the RPC method args.
	ReadRequest(interface{}) error
	// Writes the response using the RPC method reply.
	WriteResponse(http.ResponseWriter, interface{})
	// Writes an error produced by the server.
	WriteError(w http.ResponseWriter, status int, err error)
}
