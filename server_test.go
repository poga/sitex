package main

import (
	"net/http"
	"testing"

	"io/ioutil"

	"net"

	"github.com/stretchr/testify/assert"
)

func TestExampleServer(t *testing.T) {
	server, err := NewServer("./example")
	assert.NoError(t, err)
	listener, err := net.Listen("tcp", ":9069")
	assert.NoError(t, err)
	go server.Start(listener)

	resp, err := sendReq("GET", "http://localhost:9069/test.json")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"))
	body, err := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "{\"foo\": \"bar\"}\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "{\"foo\": \"bar\"}\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/foo")
	assert.NoError(t, err)
	assert.Equal(t, 301, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))
	assert.Equal(t, "/test.json", resp.Header.Get("Location"))

	resp, err = sendReq("GET", "http://localhost:9069/bar?id=2")
	assert.NoError(t, err)
	assert.Equal(t, 301, resp.StatusCode)
	assert.Equal(t, "/test-2.json", resp.Header.Get("Location"))

	resp, err = sendReq("GET", "http://localhost:9069/test-2.json")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "SiteXID", resp.Header.Get("X-TEST-HEADER"))
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "{\"foo\": \"bar2\"}\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/google")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))

	resp, err = sendReq("GET", "http://localhost:9069/secret.json")
	assert.Equal(t, "", resp.Header.Get("Basic-Auth"))
	assert.Equal(t, 401, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "Unauthorized.\n", string(body))

	resp, err = sendReqAuth("GET", "http://localhost:9069/secret.json", "user", "pass")
	assert.Equal(t, "", resp.Header.Get("Basic-Auth"))
	assert.Equal(t, 200, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "{\n  \"secret\": true\n}", string(body))

	resp, err = sendReqAuth("GET", "http://localhost:9069/secret.json", "foo", "pass")
	assert.Equal(t, "", resp.Header.Get("Basic-Auth"))
	assert.Equal(t, 401, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, "Unauthorized.\n", string(body))
}

func sendReq(method string, url string) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, nil)
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return client.Do(req)
}

func sendReqAuth(method string, url string, user string, pass string) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, nil)
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req.SetBasicAuth(user, pass)
	return client.Do(req)
}
