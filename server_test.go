package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"io/ioutil"

	"io"
	"net"
)

func TestExampleServer(t *testing.T) {
	mockProxy()

	server, err := NewServer("./example")
	require.NoError(t, err)
	listener, err := net.Listen("tcp", ":9069")
	require.NoError(t, err)
	go server.Start(listener)

	resp, err := sendReq("GET", "http://localhost:9069/test.json")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	require.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"))
	body, _ := ioutil.ReadAll(resp.Body)
	require.Equal(t, "{\"foo\": \"bar\"}\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "{\"foo\": \"bar\"}\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/notFound.json")
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)

	resp, err = sendReq("GET", "http://localhost:9069/foo")
	require.NoError(t, err)
	require.Equal(t, 301, resp.StatusCode)
	require.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))
	require.Equal(t, "/test.json", resp.Header.Get("Location"))

	resp, err = sendReq("GET", "http://localhost:9069/bar?id=2")
	require.NoError(t, err)
	require.Equal(t, 301, resp.StatusCode)
	require.Equal(t, "/test-2.json", resp.Header.Get("Location"))

	resp, err = sendReq("GET", "http://localhost:9069/bar?page=2")
	require.NoError(t, err)
	require.Equal(t, 301, resp.StatusCode)
	require.Equal(t, "/test-2.json", resp.Header.Get("Location"))

	resp, err = sendReq("GET", "http://localhost:9069/bar")
	require.NoError(t, err)
	require.Equal(t, 301, resp.StatusCode)
	require.Equal(t, "/test.json", resp.Header.Get("Location"))

	resp, err = sendReq("GET", "http://localhost:9069/test-2.json")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "SiteXID", resp.Header.Get("X-TEST-HEADER"))
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "{\"foo\": \"bar2\"}\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/google")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "proxy!", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/secret.json")
	require.Equal(t, "", resp.Header.Get("Basic-Auth"))
	require.Equal(t, 401, resp.StatusCode)
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "Unauthorized.\n", string(body))

	resp, err = sendReqAuth("GET", "http://localhost:9069/secret.json", "user", "pass")
	require.Equal(t, "", resp.Header.Get("Basic-Auth"))
	require.Equal(t, 200, resp.StatusCode)
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "{\n  \"secret\": true\n}", string(body))

	resp, err = sendReqAuth("GET", "http://localhost:9069/secret.json", "foo", "pass")
	require.Equal(t, "", resp.Header.Get("Basic-Auth"))
	require.Equal(t, 401, resp.StatusCode)
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "Unauthorized.\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/shadowed.json")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "{\"foo\": \"bar\"}\n", string(body))

	resp, err = sendReq("GET", "http://localhost:9069/notShadowed.json")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "", resp.Header.Get("X-TEST-HEADER"))
	body, _ = ioutil.ReadAll(resp.Body)
	require.Equal(t, "{\n  \"shadowed\": false\n}", string(body))
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

func mockProxy() {
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "proxy!")
	})
	go http.ListenAndServe(":9090", nil)

}
