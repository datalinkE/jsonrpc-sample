package rpcserver

import (
	"errors"
	"github.com/gin-gonic/gin"
	jsonrpc "github.com/gorilla/rpc/v2/json2"
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
	server, err := NewServer(mock)
	require.NoError(t, err)
	server.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	engine := gin.Default()
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
	require.Equal(t, 400, w.Code)
	require.Equal(t, 0, mock.Called)
}

func Test_02_GarbageBody(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `wtf`)
	require.Equal(t, 400, w.Code)
	require.Equal(t, 0, mock.Called)
}

func Test_03_InvalidJSONBody(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{}`) // no "jsonrpc"
	require.Equal(t, 400, w.Code)
	require.Equal(t, 0, mock.Called)
}

func Test_04_MissingMethodField(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0"}`)
	require.Equal(t, 400, w.Code)
	require.Equal(t, 0, mock.Called)
}

func Test_05_MissingMethodHandle(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Wrong", `{"jsonrpc": "2.0", "method": "Action", "id":1, "params": {"A": 5, "B": 2}}`)

	require.Equal(t, 404, w.Code)
	require.Equal(t, 0, mock.Called)
}

func Test_05_WrongMethodField(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Wrong", "id":1, "params": {"A": 5, "B": 2}}`)

	require.Equal(t, 404, w.Code) // TODO: maybe 400 here?
	require.Equal(t, 0, mock.Called)
}

func Test_06_WrongMethodHandleAndField(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Wrong", `{"jsonrpc": "2.0", "method": "Wrong", "id":1, "params": {"A": 5, "B": 2}}`)

	require.Equal(t, 404, w.Code)
	require.Equal(t, 0, mock.Called)
}

func Test_07_Error(t *testing.T) {
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Action", "id":1, "params": {"A": 5, "B": 5}}`) // A==B, expecting error

	body := ShowResponse(t, w)

	require.Equal(t, 200, w.Code)
	require.True(t, len(body) > 0)

	require.Equal(t, 1, mock.Called)
	require.Error(t, mock.Err)
	strings.Contains(body, "expected error A==B")
}

func Test_08_NotifyRequestSanity(t *testing.T) { // Notify request == without id
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Action", "params": {"A": 5, "B": 2}}`)

	body := ShowResponse(t, w)

	require.Equal(t, 200, w.Code)
	require.True(t, len(body) == 0) // without response body
	require.Equal(t, 1, mock.Called)
	require.Equal(t, 5, mock.A)
	require.Equal(t, 2, mock.B)
	require.Equal(t, 3, mock.Result)
	require.NoError(t, mock.Err)
}

func Test_09_NotifyRequestError(t *testing.T) { // Notify request == without id
	mock, w := performRequest(t, "POST", "/jsonrpc/v1/Action", `{"jsonrpc": "2.0", "method": "Action", "params": {"A": 5, "B": 5}}`) // A==B, expecting error

	body := ShowResponse(t, w)

	require.Equal(t, 200, w.Code)
	require.True(t, len(body) == 0) // without response body
	require.Equal(t, 1, mock.Called)
	require.Equal(t, 5, mock.A)
	require.Equal(t, 5, mock.B)
	require.Error(t, mock.Err)
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
			m.Err = errors.New("expected error A==B")
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
