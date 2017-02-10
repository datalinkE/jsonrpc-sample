// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2017 Andrey Pichugin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc2

import (
	"errors"
)

const (
	E_PARSE       = -32700
	E_INVALID_REQ = -32600
	E_NO_METHOD   = -32601
	E_BAD_PARAMS  = -32602
	E_INTERNAL    = -32603
	E_SERVER      = -32000
)

var ErrNullResult = errors.New("result is null")

type Error struct {
	// A Number that indicates the error type that occurred.
	Code int `json:"code"` /* required */

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */

	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data,omitempty"` /* optional */
}

func NewError(code int, msg string, data interface{}) error {
	return &Error{
		Code:    code,
		Message: msg,
		Data:    data,
	}
}

func (e *Error) Error() string {
	return e.Message
}
