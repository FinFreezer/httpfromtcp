package request

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	fmt.Println("Running request_test.")
	testNo := 1

	// Test 1: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	fmt.Printf("Running test nr. %d\n", testNo)
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	testNo += 1

	// Test 2: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	fmt.Printf("Running test nr. %d\n", testNo)
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	testNo += 1

	// Test 3: Invalid number of parts in request line
	reader = &chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 10,
	}
	fmt.Printf("Running test nr. %d\n", testNo)
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, r)
	testNo += 1

	// Test 4: Good POST Request with path
	reader = &chunkReader{
		data:            "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 10000,
	}
	fmt.Printf("Running test nr. %d\n", testNo)
	r, err = RequestFromReader(reader)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	testNo += 1

	// Test 5: Invalid method in request line
	reader = &chunkReader{
		data:            "POST69 /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 5,
	}
	fmt.Printf("Running test nr. %d\n", testNo)
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, r)
	/*assert.Equal(t, "POST69", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)*/
	testNo += 1

	// Test 6: Invalid version in request line
	reader = &chunkReader{
		data:            "GET /coffee HTTP/3.3\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 15,
	}
	fmt.Printf("Running test nr. %d\n", testNo)
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, r)
	/*assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "3.3", r.RequestLine.HttpVersion)*/
	testNo += 1

	// Test 7: Standard Headers
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])
	testNo += 1

	// Test 8 : Malformed Header
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	testNo += 1

	// Test 9: Empty Headers
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost:\r\nUser-Agent:\r\nAccept:\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", r.Headers["host"])
	assert.Equal(t, "", r.Headers["user-agent"])
	assert.Equal(t, "", r.Headers["accept"])
	testNo += 1

	// Test 10: No Headers
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Headers)
	testNo += 1

	// Test 11: Duplicate Headers
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: 127.0.0.1:8080\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069, 127.0.0.1:8080", r.Headers["host"])
	assert.Equal(t, "*/*", r.Headers["accept"])
	testNo += 1

	// Test 12: Missing end of headers.
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: 127.0.0.1:8080\r\nAccept: */*\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069, 127.0.0.1:8080", r.Headers["host"])
	assert.Equal(t, "*/*", r.Headers["accept"])
	testNo += 1

	// Test 13: Standard Body
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))
	testNo += 1

	// Test 14: Body shorter than reported content length
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	testNo += 1

	// Test 15: Empty Body, 0 reported content length
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	testNo += 1

	// Test 16: Empty Body, no reported content length
	fmt.Printf("Running test nr. %d\n", testNo)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	testNo += 1
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}
