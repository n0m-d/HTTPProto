package headers

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestHeader(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid multiple headers in single request
	headers = NewHeaders()
	data = []byte("Test: Test1\r\nTest: Test2\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, n > 0, "should have read some bytes")
	assert.True(t, done, "should be done parsing")
	assert.Equal(t, "Test1, Test2", headers.Get("Test"))

	// Test: A header with multiple values
	headers = NewHeaders()
	headers["set-person"] = "lane-loves-go"
	data = []byte("Set-Person: prime-loves-zig\r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.NoError(t, err)
	assert.Equal(t, 31, n)
	assert.True(t, done)

	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
}
