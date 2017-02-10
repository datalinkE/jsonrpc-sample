package main

import (
	"errors"
	"github.com/datalinkE/rpcserver"
	"github.com/datalinkE/rpcserver/jsonrpc2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ShowResponse(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Logf("resp.Code = %v", w.Code)
	t.Logf("resp.Headers:")
	for key, value := range w.Header() {
		t.Logf("'%v: %v'", key, value)
	}
	body, err := ioutil.ReadAll(w.Body)
	require.NoError(t, err)
	result := string(body)
	t.Logf("resp.Body = %v", result)
	return result
}

func performRequest(t *testing.T, getOrPost string, path string, body string) (*MockRpcObject, *httptest.ResponseRecorder) {
	mock := NewMockRpcObject(t)
	server, err := rpcserver.NewServer(mock)
	require.NoError(t, err)
	server.RegisterCodec(jsonrpc2.NewCodec(), "application/json")
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.POST("/jsonrpc/v1/:method", gin.WrapH(server))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(getOrPost, path, strings.NewReader(body))
	engine.ServeHTTP(w, req)
	return mock, w
}

func Test_01_Sanity(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Action", "id":1, "params": {"A": 5, "B": 2}}`)

	body := ShowResponse(t, w)

	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)

	require.Equal(t, 1, mock.Called)
	require.Equal(t, 5, mock.A)
	require.Equal(t, 2, mock.B)
	require.Equal(t, 3, mock.Result)
	require.NoError(t, mock.Err)
}

func Test_02_EmptyBody(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", ``)
	body := ShowResponse(t, w)
	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)
	require.True(t, strings.Contains(body, `"error"`))
	require.True(t, strings.Contains(body, `"code":-32700`)) // parse error
	require.Equal(t, 0, mock.Called)
}

func Test_02_GarbageBody(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `wtf`)
	body := ShowResponse(t, w)
	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)
	require.True(t, strings.Contains(body, `"error"`))
	require.True(t, strings.Contains(body, `"code":-32700`)) // parse error
	require.Equal(t, 0, mock.Called)
}

func Test_03_InvalidJSONBody(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{}`) // no "jsonrpc"
	body := ShowResponse(t, w)
	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)
	require.True(t, strings.Contains(body, `"error"`))
	require.True(t, strings.Contains(body, `"code":-32600`)) // invalid req error
	require.Equal(t, 0, mock.Called)
}

func Test_04_MissingMethodField(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0"}`)
	body := ShowResponse(t, w)
	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)
	require.True(t, strings.Contains(body, `"error"`))
	require.True(t, strings.Contains(body, `"code":-32601`)) // missing method error
	require.Equal(t, 0, mock.Called)
}

func Test_05_MissingMethodHandle(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Wrong", `{"jsonrpc": "2.0", "method": "Action", "id":1, "params": {"A": 5, "B": 2}}`)
	body := ShowResponse(t, w)
	require.Equal(t, 404, w.Code) // like response by server itself
	require.True(t, len(body) > 0)
	require.Equal(t, 0, mock.Called)
}

func Test_05_WrongMethodField(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Wrong", "id":1, "params": {"A": 5, "B": 2}}`)
	body := ShowResponse(t, w)
	require.Equal(t, 200, w.Code) // response by jsonrpc
	require.True(t, len(body) > 0)
	require.True(t, strings.Contains(body, `"error"`))
	require.True(t, strings.Contains(body, `"code":-32601`)) // missing method error
	require.Equal(t, 0, mock.Called)
}

func Test_06_WrongMethodHandleAndField(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Wrong", `{"jsonrpc": "2.0", "method": "Wrong", "id":1, "params": {"A": 5, "B": 2}}`)

	require.Equal(t, 404, w.Code)
	require.Equal(t, 0, mock.Called)
}

func Test_07_ErrorAny(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Action", "id":1, "params": {"A": 5, "B": 5}}`) // A==B, expecting error

	body := ShowResponse(t, w)

	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)

	require.Equal(t, 1, mock.Called)
	require.Error(t, mock.Err)
	strings.Contains(body, "expected error A==B")
}

func Test_08_ErrorSpecific(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Action", "id":1, "params": {"A": 10, "B": 1}}`) // A==Bx10, expecting specific error

	body := ShowResponse(t, w)

	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)

	require.Equal(t, 1, mock.Called)
	require.Error(t, mock.Err)
	strings.Contains(body, `{"jsonrpc":"2.0","error":{"code":500,"message":"expected error A==Bx10 - jsonrpc-aware","data":{"A":10,"B":1}},"id":1}`)
}

func Test_09_NotifyRequestHaveResponseByDefault(t *testing.T) { // Notify request == without id
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Action", "params": {"A": 5, "B": 2}}`)

	body := ShowResponse(t, w)

	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0) // without response body
	require.Equal(t, 1, mock.Called)
	require.Equal(t, 5, mock.A)
	require.Equal(t, 2, mock.B)
	require.Equal(t, 3, mock.Result)
	require.NoError(t, mock.Err)
}

type MockArgs struct {
	A, B int
}

type MockReply struct {
	Value int
}

type MockRpcObject struct {
	Called int
	A      int
	B      int
	Result int
	Err    error
	t      *testing.T
}

func NewMockRpcObject(t *testing.T) *MockRpcObject {
	return &MockRpcObject{t: t}
}

func (m *MockRpcObject) Action(r *http.Request, args *MockArgs, reply *MockReply) error {
	defer m.t.Log("Action-end")
	m.Called = m.Called + 1
	if args != nil {
		m.t.Logf("Action: args='%v'", *args)
		m.A = args.A
		m.B = args.B
		if args.A == args.B {
			m.Err = errors.New("expected error A==B - simple")
			return m.Err
		}
		if args.A == (args.B * 10) {
			m.Err = jsonrpc2.NewError(500, "expected error A==Bx10 - jsonrpc-aware", args)
			return m.Err
		}
		m.Result = args.A - args.B
	} else {
		m.t.Logf("Action: args=nil")
	}

	if reply != nil {
		reply.Value = m.Result
		m.t.Logf("Action: result='%v'", *reply)
	} else {
		m.t.Log("Action: result=nil")
	}
	return nil
}
