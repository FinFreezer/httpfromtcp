package response

import (
	"fmt"
	"strings"
	"testing"

	"github.com/finfreezer/httpfromtcp/internal/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLResponse(t *testing.T) {
	newCr := chunkWriter{data: []byte{}, pos: 0}
	newWriter := Writer{ContentWriter: &newCr, ResponseHTML: respTemplate, Headers: headers.NewHeaders()}
	fmt.Println("Running response_test.")
	testNo := 1

	// Test 1: Replace status code.
	fmt.Printf("Running test nr. %d\n", testNo)
	newStatus := 502
	err := newWriter.WriteStatusLine(StatusCode(newStatus))
	require.NoError(t, err)
	assert.Equal(t, true, strings.Contains(newWriter.ResponseHTML, "<title>502"))
	fmt.Println(newWriter.ResponseHTML)
	testNo += 1

	// Test 2: Replace status code.
	fmt.Printf("Running test nr. %d\n", testNo)
	newStatus = 400
	err = newWriter.WriteStatusLine(StatusCode(newStatus))
	require.NoError(t, err)
	assert.Equal(t, true, strings.Contains(newWriter.ResponseHTML, "<title>400 Bad Request"))
	fmt.Println(newWriter.ResponseHTML)
	testNo += 1

	// Test 3: Replace body.
	fmt.Printf("Running test nr. %d\n", testNo)
	newBody := []byte("This has been a wonderful experiment.")
	n, err := newWriter.WriteBody(newBody)
	require.NoError(t, err)
	assert.Equal(t, true, strings.Contains(newWriter.ResponseHTML, "<p>This"))
	assert.Equal(t, len(newBody), n)
	fmt.Println(newWriter.ResponseHTML)
	testNo += 1
}

type chunkWriter struct {
	data []byte
	pos  int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkWriter) Write(p []byte) (n int, err error) {
	n = copy(p, cr.data[cr.pos:])
	cr.pos += n

	return n, nil
}
