package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestParseHeaderComment(t *testing.T) {
	_, err := NewHeaderRouter([]byte("# just a comment"))
	assert.NoError(t, err)
}

func TestParseHeader(t *testing.T) {
	config := `
/foo
	X-TEST-HEADER: bar
	`
	router, err := NewHeaderRouter([]byte(config))
	assert.NoError(t, err)
	handle, params, _ := router.Lookup("GET", "/foo")
	assert.NotNil(t, handle)
	assert.Nil(t, params)
}
