package main

import (
	"net/http"
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

func TestExampleServer(t *testing.T) {
	server, err := NewServer("./example", ":9069")
	assert.NoError(t, err)
	go server.Start()

	resp, err := sendReq("GET", "http://localhost:9069/test.json")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "SiteX", resp.Header.Get("X-TEST-HEADER"))
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
