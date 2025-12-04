package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test:Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	host, _ := headers.Get("Host")
	assert.Equal(t, "localhost:42069", host)
	missingKey, _ := headers.Get("Missing Key")
	assert.Equal(t, "", missingKey)
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test:Valid 2 headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nFoo: Bar\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	host, _ = headers.Get("Host")
	foo, _ := headers.Get("Foo")
	assert.Equal(t, "localhost:42069", host)
	assert.Equal(t, "Bar", foo)
	assert.Equal(t, 35, n)
	assert.True(t, done)

	// Test: Header with multiple values
	headers = NewHeaders()
	data = []byte("Person: Frank\r\nPerson: Gojo\r\nPerson: Yuji\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	person, _ := headers.Get("Person")
	assert.Equal(t, "Frank, Gojo, Yuji", person)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("	 Host : localhost:42069		\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character in field name
	headers = NewHeaders()
	data = []byte("	 H😊st: localhost:42069		\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}