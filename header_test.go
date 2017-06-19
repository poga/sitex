package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseHeaderComment(t *testing.T) {
	_, err := NewHeaders([]byte("# just a comment"))
	require.NoError(t, err)
}

func TestParseHeader(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestParseIncompleteHeader(t *testing.T) {
	config := `
/foo
	`
	routers, err := NewHeaders([]byte(config))
	require.Error(t, err)
	require.Nil(t, routers)
}

func TestParseIncorrectIndentedHeaderPath(t *testing.T) {
	config := `
	/foo
	X-TEST-HEADER: bar
	`
	routers, err := NewHeaders([]byte(config))
	require.Error(t, err)
	require.Nil(t, routers)
}

func TestParseIncorrectIndentedHeaderHeader(t *testing.T) {
	config := `
/foo
X-TEST-HEADER: bar
	`
	routers, err := NewHeaders([]byte(config))
	require.Error(t, err)
	require.Nil(t, routers)
}

func TestParseUnclosedPath(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
/bar
	`
	routers, err := NewHeaders([]byte(config))
	require.Error(t, err)
	require.Nil(t, routers)
}

func TestParseHeaderWithEmptyLine(t *testing.T) {
	config := `
/foo

	X-TEST-HEADER: bar
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestParseHeaderWithEmptyLine2(t *testing.T) {
	config := `

/foo
	X-TEST-HEADER: bar
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestParseHeaderWithWhitespace(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar baz
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar baz", res.Header().Get("X-TEST-HEADER"))
}

func TestParseHeaderIncludeColon(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar:baz
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar:baz", res.Header().Get("X-TEST-HEADER"))
}

func TestParseHeaderWithInlineComment(t *testing.T) {
	config := `
/foo # hi
	X-TEST-HEADER: bar #hello
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestMultiKeyHeader(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
	X-TEST-HEADER: baz
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar, baz", res.Header().Get("X-TEST-HEADER"))
}

func TestMultiHeader(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
	X-TEST-HEADER2: baz
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, "baz", res.Header().Get("X-TEST-HEADER2"))
}

func TestPathMatchingSplat(t *testing.T) {
	config := `
/*
	X-TEST-HEADER: bar
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestPathMatchingSplatWithPrefix(t *testing.T) {
	config := `
/prefix/*
	X-TEST-HEADER: bar
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/prefix/foo", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestPathMatchingPlaceholder(t *testing.T) {
	config := `
/:foo/bar
	X-TEST-HEADER: bar
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/abc/bar", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
}

func TestMatchMultiplePath(t *testing.T) {
	config := `
/bar/:x
	X-TEST-HEADER: bar
/baz
	X-TEST: baz
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/bar/abc", nil)
	require.True(t, headers[0].Match(req))
	req, _ = http.NewRequest("GET", "/bar/abc", nil)
	require.False(t, headers[1].Match(req))
	req, _ = http.NewRequest("GET", "/baz", nil)
	require.False(t, headers[0].Match(req))
	req, _ = http.NewRequest("GET", "/baz", nil)
	require.True(t, headers[1].Match(req))

	req, _ = http.NewRequest("GET", "/bar/abc", nil)
	res := testHeader(headers[0], req)
	require.Equal(t, "bar", res.Header().Get("X-TEST-HEADER"))
	req, _ = http.NewRequest("GET", "/baz", nil)
	res = testHeader(headers[1], req)
	require.Equal(t, "baz", res.Header().Get("X-TEST"))
}

func TestPathBasicAuth(t *testing.T) {
	config := `
/login
	Basic-Auth: foo:bar
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/login", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)

	res = testHeaderAuth(headers[0], req, "foo", "bar")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(headers[0], req, "foo", "baz")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)
}

func TestPathDuplicatedBasicAuth(t *testing.T) {
	config := `
/login
	Basic-Auth: foo:bar
	Basic-Auth: foo:baz
	`
	routers, err := NewHeaders([]byte(config))
	require.Error(t, err)
	require.Nil(t, routers)
}
func TestPathMultipleBasicAuth(t *testing.T) {
	config := `
/login
	Basic-Auth: foo:bar aaa:bbb
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/login", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)

	res = testHeaderAuth(headers[0], req, "foo", "bar")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(headers[0], req, "aaa", "bbb")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(headers[0], req, "foo", "baz")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, 401, res.Code)
}

func TestPathBasicAuthAndHeader(t *testing.T) {
	config := `
/login
	Basic-Auth: foo:bar
	X-TEST-HEADER: hello
	`
	headers, err := NewHeaders([]byte(config))
	require.NoError(t, err)
	req, _ := http.NewRequest("GET", "/login", nil)
	require.True(t, headers[0].Match(req))

	res := testHeader(headers[0], req)
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, "hello", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, 401, res.Code)

	res = testHeaderAuth(headers[0], req, "foo", "bar")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, "hello", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, 200, res.Code)

	res = testHeaderAuth(headers[0], req, "foo", "baz")
	require.Equal(t, "", res.Header().Get("Basic-Auth"))
	require.Equal(t, "hello", res.Header().Get("X-TEST-HEADER"))
	require.Equal(t, 401, res.Code)
}

func testHeader(header Header, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	header.Handle(rec, req)

	return rec
}

func testHeaderAuth(header Header, req *http.Request, username string, password string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req.SetBasicAuth(username, password)
	header.Handle(rec, req)

	return rec
}
