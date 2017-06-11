package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestParseComment(t *testing.T) {
	route, err := ParseRedirectRule([]byte("# This is a comment"))
	assert.NoError(t, err)
	assert.Nil(t, route)
}

func TestParseEmptyLine(t *testing.T) {
	route, err := ParseRedirectRule([]byte("    "))
	assert.NoError(t, err)
	assert.Nil(t, route)
}

func TestParseBasicRule(t *testing.T) {
	route, err := ParseRedirectRule([]byte("/ /foo"))
	assert.NoError(t, err)
	assert.Equal(t, 301, route.StatusCode)
	assert.Equal(t, "/", route.Match)
	assert.Equal(t, "/foo", route.To)

	resp := testRequest(route, "GET", "/")
	assert.Equal(t, 301, resp.Code)
	assert.Equal(t, "/foo", resp.HeaderMap["Location"][0])
}

func TestParseStatusCode(t *testing.T) {
	route, err := ParseRedirectRule([]byte("/ /test/test.json 200"))
	assert.NoError(t, err)
	assert.Equal(t, 200, route.StatusCode)
	assert.Equal(t, "/", route.Match)
	assert.Equal(t, "/test/test.json", route.To)

	resp := testRequest(route, "GET", "/")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "{\"foo\": \"bar\"}\n", resp.Body.String())
}

func TestParseInvalidStatusCode(t *testing.T) {
	_, err := ParseRedirectRule([]byte("/ /foo bar"))
	assert.Error(t, err)
}
func TestParsePlaceholderRule(t *testing.T) {
	route, err := ParseRedirectRule([]byte("/news/:year /foo/:year"))
	assert.NoError(t, err)
	assert.Equal(t, 301, route.StatusCode)
	assert.Equal(t, "/news/:year", route.Match)
	assert.Equal(t, "/foo/:year", route.To)
}

func TestParseSplatRule(t *testing.T) {
	route, err := ParseRedirectRule([]byte("/news/* /:splat"))
	assert.NoError(t, err)
	assert.Equal(t, 301, route.StatusCode)
	assert.Equal(t, "/news/*splat", route.Match)
	assert.Equal(t, "/:splat", route.To)

	resp := testRequest(route, "GET", "/news/foo")
	assert.Equal(t, 301, resp.Code)
	assert.Equal(t, "/foo", resp.HeaderMap["Location"][0])

	resp = testRequest(route, "GET", "/news/test/test.json")
	assert.Equal(t, 301, resp.Code)
	assert.Equal(t, "/test/test.json", resp.HeaderMap["Location"][0])

}

func TestParseQueryParams(t *testing.T) {
	route, err := ParseRedirectRule([]byte("/test/test.json id=:id  /foo/:id  301"))
	assert.NoError(t, err)
	assert.Equal(t, 301, route.StatusCode)
	assert.Equal(t, "/test/test.json", route.Match)
	assert.Equal(t, "/foo/:id", route.To)
	assert.Equal(t, "id", route.Queries["id"])

	resp := testRequest(route, "GET", "/test/test.json")
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "{\"foo\": \"bar\"}\n", resp.Body.String())

	resp = testRequest(route, "GET", "/test/test.json?id=1")
	assert.Equal(t, 301, resp.Code)
	assert.Equal(t, "/foo/1", resp.HeaderMap["Location"][0])
}

func TestParseExcessiveFields(t *testing.T) {
	_, err := ParseRedirectRule([]byte("/store id=:id  /blog/:id  301 foo"))
	assert.Error(t, err)
}

func testRequest(route *Route, method string, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	router := httprouter.New()
	router.Handle(method, route.Match, route.Handler)
	router.ServeHTTP(rec, req)

	return rec
}
