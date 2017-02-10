// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2017 Andrey Pichugin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc2

import (
	"encoding/json"
	"fmt"
	"github.com/datalinkE/rpcserver"
	"net/http"
)

var null = json.RawMessage([]byte("null"))
var Version = "2.0"

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// serverRequest represents a JSON-RPC request received by the server.
type serverRequest struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`

	// The request id. MUST be a string, number or null.
	// Our implementation will not do type checking for id.
	// It will be copied as it is.
	Id *json.RawMessage `json:"id"`
}

// serverResponse represents a JSON-RPC response returned by the server.
type serverResponse struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	// As per spec the member will be omitted if there was an error.
	Result interface{} `json:"result,omitempty"`

	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	// As per spec the member will be omitted if there was no error.
	Error *Error `json:"error,omitempty"`

	// This must be the same id as the request it is responding to.
	Id *json.RawMessage `json:"id,omitempty"`
}

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// Codec creates a CodecRequest to process each request.
type Codec struct {
	RespectNotifyMessages bool
}

// NewCodec creates a Codec object.
func NewCodec() *Codec {
	return &Codec{
		RespectNotifyMessages: false,
	}
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

// NewRequest returns a CodecRequest. Decode the request body and check if RPC signature is valid.
func (c *Codec) NewRequest(r *http.Request) rpcserver.CodecRequest {
	req := new(serverRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		err = NewError(E_PARSE, err.Error(), req)
	} else if req.Version != Version {
		err = NewError(E_INVALID_REQ, "jsonrpc must be "+Version, req)
	} else if req.Method == "" {
		err = NewError(E_NO_METHOD, "method field empty or missing", req)
	} else {
		pathMethod := rpcserver.LastPart(r.URL.Path)
		if pathMethod != req.Method {
			err = NewError(E_NO_METHOD, fmt.Sprintf("rpc: URL.Path '%v' does not end with method Name '%v'", r.URL.Path, req.Method), req)
		}
	}
	r.Body.Close()
	return &CodecRequest{request: req, err: err, respectNotifyMessages: c.RespectNotifyMessages}
}

// CodecRequest decodes and encodes a single request.
type CodecRequest struct {
	request               *serverRequest
	err                   error
	respectNotifyMessages bool
}

// Error returns if request was valid or incorrect.
func (c *CodecRequest) Error() error {
	return c.err
}

// Method returns the RPC method for the current request.
func (c *CodecRequest) Method() (string, error) {
	if c.err == nil {
		return c.request.Method, nil
	}
	return "", c.err
}

// ReadRequest fills the request object for the RPC method.
//
// ReadRequest parses request parameters in two supported forms in
// accordance with http://www.jsonrpc.org/specification#parameter_structures
//
// by-position: params MUST be an Array, containing the
// values in the Server expected order.
//
// by-name: params MUST be an Object, with member names
// that match the Server expected parameter names. The
// absence of expected names MAY result in an error being
// generated. The names MUST match exactly, including
// case, to the method's expected parameters.
func (c *CodecRequest) ReadRequest(args interface{}) error {
	if c.err == nil && c.request.Params != nil {
		// Note: if c.request.Params is nil it's not an error, it's an optional member.
		// JSON params structured object. Unmarshal to the args object.
		if err := json.Unmarshal(*c.request.Params, args); err != nil {
			// Clearly JSON params is not a structured object,
			// fallback and attempt an unmarshal with JSON params as
			// array value and RPC params is struct. Unmarshal into
			// array containing the request struct.
			params := [1]interface{}{args}
			if err = json.Unmarshal(*c.request.Params, &params); err != nil {
				c.err = &Error{
					Code:    E_INVALID_REQ,
					Message: err.Error(),
					Data:    c.request.Params,
				}
			}
		}
	}
	return c.err
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
func (c *CodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}) {
	res := &serverResponse{
		Version: Version,
		Result:  reply,
		Id:      c.request.Id,
	}
	c.writeServerResponse(w, res)
}

func (c *CodecRequest) WriteError(w http.ResponseWriter, status int, err error) {
	jsonErr, ok := err.(*Error)
	if !ok {
		jsonErr = &Error{
			Code:    status,
			Message: err.Error(),
		}
	}
	res := &serverResponse{
		Version: Version,
		Error:   jsonErr,
		Id:      c.request.Id,
	}
	c.writeServerResponse(w, res)
}

func (c *CodecRequest) writeServerResponse(w http.ResponseWriter, res *serverResponse) {
	// Id is null for notifications and they don't have a response.
	if c.request.Id == nil && c.respectNotifyMessages {
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	err := encoder.Encode(res)

	// Not sure in which case will this happen. But seems harmless.
	if err != nil {
		rpcserver.WriteError(w, 400, err.Error())
	}
}
