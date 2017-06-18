package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseHeaderComment(t *testing.T) {
	_, err := NewHeaderRouters([]byte("# just a comment"))
	require.NoError(t, err)
}

func TestParseHeader(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
	`
	routers, err := NewHeaderRouters([]byte(config))
	require.NoError(t, err)
	handle, params, _ := routers[0].Lookup("GET", "/foo")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(routers[0], "GET", "/foo")
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestParseHeaderWithWhitespace(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar baz
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/foo")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/foo")
	require.Equal(t, "bar baz", res.Header().Get("X-TEST-HEADER"))
}

func TestParseHeaderIncludeColon(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar:baz
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/foo")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/foo")
	require.Equal(t, "bar:baz", res.Header().Get("X-TEST-HEADER"))
}

func TestParseHeaderWithInlineComment(t *testing.T) {
	config := `
/foo # hi
	X-TEST-HEADER: bar #hello
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/foo")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/foo")
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestMultiKeyHeader(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
	X-TEST-HEADER: baz
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/foo")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/foo")
	require.Equal(t, "bar, baz", res.Header().Get("X-TEST-HEADER"))
}

func TestMultiHeader(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
	X-TEST-HEADER2: baz
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/foo")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/foo")
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, "baz", res.Header().Get("X-TEST-HEADER2"))
}

func TestPathMatchingSplat(t *testing.T) {
	config := `
/*
	X-TEST-HEADER: bar
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/foo")
	require.NotNil(t, handle)
	require.NotNil(t, params)

	res := testHeader(router, "GET", "/foo")
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestPathMatchingSplatWithPrefix(t *testing.T) {
	config := `
/prefix/*
	X-TEST-HEADER: bar
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/prefix/foo")
	require.NotNil(t, handle)
	require.NotNil(t, params)

	res := testHeader(router, "GET", "/prefix/foo")
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestPathMatchingPlaceholder(t *testing.T) {
	config := `
/:foo/bar
	X-TEST-HEADER: bar
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/abc/bar")
	require.NotNil(t, handle)
	require.NotNil(t, params)

	res := testHeader(router, "GET", "/abc/bar")
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestMatchMultiplePath(t *testing.T) {
	config := `
/bar/:x
	X-TEST-HEADER: bar
/baz
	X-TEST: baz
	`
	routers, err := NewHeaderRouters([]byte(config))
	require.NoError(t, err)
	handle, params, _ := routers[0].Lookup("GET", "/bar/abc")
	require.NotNil(t, handle)
	require.NotNil(t, params)
	handle, params, _ = routers[1].Lookup("GET", "/bar/abc")
	require.Nil(t, handle)
	require.Nil(t, params)
	handle, params, _ = routers[0].Lookup("GET", "/baz")
	require.Nil(t, handle)
	require.Nil(t, params)
	handle, params, _ = routers[1].Lookup("GET", "/baz")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(routers[0], "GET", "/bar/abc")
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
	res = testHeader(routers[1], "GET", "/baz")
	require.Equal(t, "baz", res.Header().Get("X-TEST"))
}

func TestPathBasicAuth(t *testing.T) {
	config := `
/login
	Basic-Auth: foo:bar
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/login")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/login")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)

	res = testHeaderAuth(router, "GET", "/login", "foo", "bar")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(router, "GET", "/login", "foo", "baz")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)
}

func TestPathMultipleBasicAuth(t *testing.T) {
	config := `
/login
	Basic-Auth: foo:bar aaa:bbb
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/login")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/login")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)

	res = testHeaderAuth(router, "GET", "/login", "foo", "bar")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(router, "GET", "/login", "aaa", "bbb")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(router, "GET", "/login", "foo", "baz")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)
}

func TestPathBasicAuthAndHeader(t *testing.T) {
	config := `
/login
	Basic-Auth: foo:bar
	X-TEST-HEADER: hello
	`
	routers, err := NewHeaderRouters([]byte(config))
	router := routers[0]
	require.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/login")
	require.NotNil(t, handle)
	require.Nil(t, params)

	res := testHeader(router, "GET", "/login")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, "hello", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, 401, res.Code)

	res = testHeaderAuth(router, "GET", "/login", "foo", "bar")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, "hello", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(router, "GET", "/login", "foo", "baz")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, "hello", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, 401, res.Code)
}

func testHeader(router HeaderRouter, method string, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	router.ServeHTTP(rec, req)

	return rec
}

func testHeaderAuth(router HeaderRouter, method string, path string, username string, password string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	req.SetBasicAuth(username, password)
	router.ServeHTTP(rec, req)

	return rec
}
