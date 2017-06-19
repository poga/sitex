package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	"github.com/stretchr/testify/require"
)

func TestParseComment(t *testing.T) {
	route, err := NewRedirect(".", []byte("# This is a comment"))
	require.NoError(t, err)
	require.Nil(t, route)
}

func TestParseEmptyLine(t *testing.T) {
	route, err := NewRedirect(".", []byte("    "))
	require.NoError(t, err)
	require.Nil(t, route)
}

func TestParseBasicRule(t *testing.T) {
	route, err := NewRedirect(".", []byte("/ /foo"))
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/", route.From)
	require.Equal(t, "/foo", route.To)

	req, _ := http.NewRequest("GET", "/", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo", resp.HeaderMap["Location"][0])

	req, _ = http.NewRequest("GET", "/foo", nil)
	require.False(t, route.Match(req))
}

func TestParseInlineComment(t *testing.T) {
	route, err := NewRedirect(".", []byte("/ /foo #hi"))
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/", route.From)
	require.Equal(t, "/foo", route.To)

	req, _ := http.NewRequest("GET", "/", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo", resp.HeaderMap["Location"][0])
}

func TestParseStatusCode(t *testing.T) {
	route, err := NewRedirect(".", []byte("/ /example/test.json 200"))
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.From)
	require.Equal(t, "/example/test.json", route.To)

	req, _ := http.NewRequest("GET", "/", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", resp.Body.String())
}

func TestParseShadowingStatusCode(t *testing.T) {
	route, err := NewRedirect(".", []byte("/ /example/test.json 200!"))
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.From)
	require.Equal(t, "/example/test.json", route.To)
	require.True(t, route.Shadowing)

	req, _ := http.NewRequest("GET", "/", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", resp.Body.String())
}

func TestParseInvalidStatusCode(t *testing.T) {
	_, err := NewRedirect(".", []byte("/ /foo bar"))
	require.Error(t, err)
}

func TestParsePlaceholderRule(t *testing.T) {
	route, err := NewRedirect(".", []byte("/news/:year /foo/:year"))
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/news/:year", route.From)
	require.Equal(t, "/foo/:year", route.To)

	req, _ := http.NewRequest("GET", "/news/2017", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo/2017", resp.HeaderMap["Location"][0])
}

func TestParsePlaceholderRuleInline(t *testing.T) {
	route, err := NewRedirect(".", []byte("/news/year-:year /foo/:year"))
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/news/year-:year", route.From)
	require.Equal(t, "/foo/:year", route.To)

	req, _ := http.NewRequest("GET", "/news/year-2017", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo/2017", resp.HeaderMap["Location"][0])
}

func TestParseSplatRule(t *testing.T) {
	route, err := NewRedirect(".", []byte("/news/* /:splat"))
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/news/*splat", route.From)
	require.Equal(t, "/:splat", route.To)

	req, _ := http.NewRequest("GET", "/news/foo", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo", resp.HeaderMap["Location"][0])

	req, _ = http.NewRequest("GET", "/news/test/test.json", nil)
	require.True(t, route.Match(req))

	resp = testRequest(route, req)
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/test/test.json", resp.HeaderMap["Location"][0])

}

func TestParseQueryParams(t *testing.T) {
	route, err := NewRedirect(".", []byte("/example/test.json id=:id  /foo/:id  301"))
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/example/test.json", route.From)
	require.Equal(t, "/foo/:id", route.To)
	require.Equal(t, "id", route.Queries["id"])

	req, _ := http.NewRequest("GET", "/example/test.json", nil)
	require.False(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "", resp.Body.String())

	req, _ = http.NewRequest("GET", "/example/test.json?id=1", nil)
	require.True(t, route.Match(req))

	resp = testRequest(route, req)
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo/1", resp.HeaderMap["Location"][0])
}

func TestParseProxy(t *testing.T) {
	ts := mockServer()
	defer ts.Close()
	route, err := NewRedirect(".", []byte(fmt.Sprintf("/  %s 200", ts.URL)))
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.From)
	require.Equal(t, ts.URL, route.To)

	req, _ := http.NewRequest("GET", "/", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "METHOD: GET", resp.Body.String())
}

func TestParseProxyPOST(t *testing.T) {
	ts := mockServer()
	defer ts.Close()
	route, err := NewRedirect(".", []byte(fmt.Sprintf("/ %s  200", ts.URL)))
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.From)
	require.Equal(t, ts.URL, route.To)

	req, _ := http.NewRequest("POST", "/", nil)
	require.True(t, route.Match(req))

	resp := testRequest(route, req)
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "METHOD: POST", resp.Body.String())
}

func TestParseExcessiveFields(t *testing.T) {
	_, err := NewRedirect(".", []byte("/store id=:id  /blog/:id  301 foo"))
	require.Error(t, err)
}

func testRequest(route *Redirect, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	route.Handle(rec, req)

	return rec
}

func mockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "METHOD: %s", r.Method)
	})
	return httptest.NewServer(mux)
}
