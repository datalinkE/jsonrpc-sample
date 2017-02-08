// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpcserver

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// Precompute the reflect.Type of error and http.Request
	TypeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	TypeOfRequest = reflect.TypeOf((*http.Request)(nil)).Elem()
)

// ----------------------------------------------------------------------------
// RpcService
// ----------------------------------------------------------------------------

type RpcService struct {
	name     string                       // name of service
	rcvr     reflect.Value                // receiver of methods for the service
	rcvrType reflect.Type                 // type of the receiver
	methods  map[string]*RpcServiceMethod // registered methods
}

type RpcServiceMethod struct {
	method    reflect.Method // receiver method
	argsType  reflect.Type   // type of the request argument
	replyType reflect.Type   // type of the response argument
}

// NewRpcService creates a RpcService object with assotiated RpcServiceMethods.
func NewRpcService(rcvr interface{}, name string) (*RpcService, error) {
	// Setup service.
	s := &RpcService{
		name:     name,
		rcvr:     reflect.ValueOf(rcvr),
		rcvrType: reflect.TypeOf(rcvr),
		methods:  make(map[string]*RpcServiceMethod),
	}
	if name == "" {
		s.name = reflect.Indirect(s.rcvr).Type().Name()
		if !IsExported(s.name) {
			return nil, fmt.Errorf("rpc: type %q is not exported", s.name)
		}
	}
	if s.name == "" {
		return nil, fmt.Errorf("rpc: no service name for type %q",
			s.rcvrType.String())
	}
	// Setup methods.
	for i := 0; i < s.rcvrType.NumMethod(); i++ {
		method := s.rcvrType.Method(i)
		mtype := method.Type
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs four ins: receiver, *http.Request, *args, *reply.
		if mtype.NumIn() != 4 {
			continue
		}
		// First argument must be a pointer and must be http.Request.
		reqType := mtype.In(1)
		if reqType.Kind() != reflect.Ptr || reqType.Elem() != TypeOfRequest {
			continue
		}
		// Second argument must be a pointer and must be exported.
		args := mtype.In(2)
		if args.Kind() != reflect.Ptr || !IsExportedOrBuiltin(args) {
			continue
		}
		// Third argument must be a pointer and must be exported.
		reply := mtype.In(3)
		if reply.Kind() != reflect.Ptr || !IsExportedOrBuiltin(reply) {
			continue
		}
		// Method needs one out: error.
		if mtype.NumOut() != 1 {
			continue
		}
		if returnType := mtype.Out(0); returnType != TypeOfError {
			continue
		}
		s.methods[method.Name] = &RpcServiceMethod{
			method:    method,
			argsType:  args.Elem(),
			replyType: reply.Elem(),
		}
	}
	if len(s.methods) == 0 {
		return nil, fmt.Errorf("rpc: %q has no exported methods of suitable type",
			s.name)
	}
	return s, nil
}

// get returns a registered service given a method name.
//
// The method name uses a dotted notation as in "Service.Method".
func (service *RpcService) Get(method string) (*RpcServiceMethod, error) {
	parts := strings.Split(method, ".")
	if len(parts) != 2 {
		err := fmt.Errorf("rpc: service/method request ill-formed: %q", method)
		return nil, err
	}
	if service == nil {
		err := fmt.Errorf("rpc: can't find service %q", method)
		return nil, err
	}
	serviceMethod := service.methods[parts[1]]
	if serviceMethod == nil {
		err := fmt.Errorf("rpc: can't find method %q", method)
		return nil, err
	}
	return serviceMethod, nil
}

// isExported returns true of a string is an exported (upper case) name.
func IsExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// isExportedOrBuiltin returns true if a type is exported or a builtin.
func IsExportedOrBuiltin(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return IsExported(t.Name()) || t.PkgPath() == ""
}
