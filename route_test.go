package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
)

func TestParseComment(t *testing.T) {
	route, err := NewRoute(".", []byte("# This is a comment"), nil)
	require.NoError(t, err)
	require.Nil(t, route)
}

func TestParseEmptyLine(t *testing.T) {
	route, err := NewRoute(".", []byte("    "), nil)
	require.NoError(t, err)
	require.Nil(t, route)
}

func TestParseBasicRule(t *testing.T) {
	route, err := NewRoute(".", []byte("/ /foo"), nil)
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/", route.Match)
	require.Equal(t, "/foo", route.To)

	resp := testRequest(route, "GET", "/")
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo", resp.HeaderMap["Location"][0])
}

func TestParseInlineComment(t *testing.T) {
	route, err := NewRoute(".", []byte("/ /foo #hi"), nil)
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/", route.Match)
	require.Equal(t, "/foo", route.To)

	resp := testRequest(route, "GET", "/")
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo", resp.HeaderMap["Location"][0])
}

func TestParseStatusCode(t *testing.T) {
	route, err := NewRoute(".", []byte("/ /test/test.json 200"), nil)
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.Match)
	require.Equal(t, "/test/test.json", route.To)

	resp := testRequest(route, "GET", "/")
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", resp.Body.String())
}

func TestParseShadowingStatusCode(t *testing.T) {
	route, err := NewRoute(".", []byte("/ /test/test.json 200!"), nil)
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.Match)
	require.Equal(t, "/test/test.json", route.To)

	resp := testRequest(route, "GET", "/")
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", resp.Body.String())
}

func TestParseInvalidStatusCode(t *testing.T) {
	_, err := NewRoute(".", []byte("/ /foo bar"), nil)
	require.Error(t, err)
}

func TestParsePlaceholderRule(t *testing.T) {
	route, err := NewRoute(".", []byte("/news/:year /foo/:year"), nil)
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/news/:year", route.Match)
	require.Equal(t, "/foo/:year", route.To)

	resp := testRequest(route, "GET", "/news/2017")
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo/2017", resp.HeaderMap["Location"][0])
}

func TestParsePlaceholderRuleInline(t *testing.T) {
	route, err := NewRoute(".", []byte("/news/year-:year /foo/:year"), nil)
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/news/year-:year", route.Match)
	require.Equal(t, "/foo/:year", route.To)

	resp := testRequest(route, "GET", "/news/year-2017")
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo/2017", resp.HeaderMap["Location"][0])
}

func TestParseSplatRule(t *testing.T) {
	route, err := NewRoute(".", []byte("/news/* /:splat"), nil)
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/news/*splat", route.Match)
	require.Equal(t, "/:splat", route.To)

	resp := testRequest(route, "GET", "/news/foo")
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo", resp.HeaderMap["Location"][0])

	resp = testRequest(route, "GET", "/news/test/test.json")
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/test/test.json", resp.HeaderMap["Location"][0])

}

func TestParseQueryParams(t *testing.T) {
	route, err := NewRoute(".", []byte("/test/test.json id=:id  /foo/:id  301"), nil)
	require.NoError(t, err)
	require.Equal(t, 301, route.StatusCode)
	require.Equal(t, "/test/test.json", route.Match)
	require.Equal(t, "/foo/:id", route.To)
	require.Equal(t, "id", route.Queries["id"])

	resp := testRequest(route, "GET", "/test/test.json")
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "{\"foo\": \"bar\"}\n", resp.Body.String())

	resp = testRequest(route, "GET", "/test/test.json?id=1")
	require.Equal(t, 301, resp.Code)
	require.Equal(t, "/foo/1", resp.HeaderMap["Location"][0])
}

func TestParseProxy(t *testing.T) {
	ts := mockServer()
	defer ts.Close()
	route, err := NewRoute(".", []byte(fmt.Sprintf("/  %s 200", ts.URL)), nil)
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.Match)
	require.Equal(t, ts.URL, route.To)

	resp := testRequest(route, "GET", "/")
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "METHOD: GET", resp.Body.String())
}

func TestParseProxyPOST(t *testing.T) {
	ts := mockServer()
	defer ts.Close()
	route, err := NewRoute(".", []byte(fmt.Sprintf("/ %s  200", ts.URL)), nil)
	require.NoError(t, err)
	require.Equal(t, 200, route.StatusCode)
	require.Equal(t, "/", route.Match)
	require.Equal(t, ts.URL, route.To)

	resp := testRequest(route, "POST", "/")
	require.Equal(t, 200, resp.Code)
	require.Equal(t, "METHOD: POST", resp.Body.String())
}
func TestParseExcessiveFields(t *testing.T) {
	_, err := NewRoute(".", []byte("/store id=:id  /blog/:id  301 foo"), nil)
	require.Error(t, err)
}

func TestRedirectWithHeaderRouter(t *testing.T) {
	headerRouters, _ := NewHeaderRouters([]byte(`
/test/test.json
	X-TEST: hello
	`))
	route, err := NewRoute(".", []byte("/test/test.json /foo/:id  301"), headerRouters)
	require.NoError(t, err)
	resp := testRequest(route, "GET", "/test/test.json")
	require.Equal(t, "hello", resp.Header().Get("X-TEST"))

	resp = testRequest(route, "GET", "/foo")
	require.Equal(t, "", resp.Header().Get("X-TEST"))
}

func testRequest(route *Route, method string, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	router := httprouter.New()
	router.Handle(method, route.Match, route.Handler)
	router.ServeHTTP(rec, req)

	return rec
}

func mockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "METHOD: %s", r.Method)
	})
	return httptest.NewServer(mux)
}
