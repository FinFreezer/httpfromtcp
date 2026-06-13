package headers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	testNo := 1

	// Test 1: Valid single header
	fmt.Printf("Running test nr. %d\n", testNo)
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
	testNo += 1

	// Test 2: Invalid spacing header
	fmt.Printf("Running test nr. %d\n", testNo)
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
	testNo += 1

	//Test 3: Multiple headers
	fmt.Printf("Running test nr. %d\n", testNo)
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n\r\n")
	n, done, err = headers.Parse(data)
	n, done, err = headers.Parse(data[n:])
	fmt.Println(headers)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "curl/7.81.0", headers["User-Agent"])
	testNo += 1

	//Test 4: Valid single header with extra whitespace
	fmt.Printf("Running test nr. %d\n", testNo)
	headers = NewHeaders()
	data = []byte("Host:       localhost:42069     \r\n\r\n")
	n, done, err = headers.Parse(data)
	fmt.Println(headers)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	testNo += 1

	//Test 5: Valid Done
	fmt.Printf("Running test nr. %d\n", testNo)
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	n, done, err = headers.Parse(data[n:])
	fmt.Println(headers)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, done, true)
	assert.Equal(t, "localhost:42069", headers["Host"])
	testNo += 1

}
