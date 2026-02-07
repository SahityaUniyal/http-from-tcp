package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders_Parse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: multiple header values
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n Second: Second\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "Second", headers.Get("Second"))
	assert.Equal(t, 42, n)
	assert.True(t, done)

	// Test: multiple header values with same key
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n Same: Same \r\n Same: Same\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "Same, Same", headers.Get("Same"))
	assert.Equal(t, 52, n)
	assert.True(t, done)

	// Test: Invalid header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, 0, n)
	require.Error(t, err)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, 0, n)
	require.Error(t, err)
	assert.False(t, done)
}
