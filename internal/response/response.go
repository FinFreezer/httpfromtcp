package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/finfreezer/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

type Writer struct {
	StatusLine []byte
	Headers    headers.Headers
	body       []byte
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case OK:
		resp := []byte("HTTP/1.1 200 OK\r\n")
		w.Write(resp)
		return nil
	case BadRequest:
		resp := []byte("HTTP/1.1 400 Bad Request\r\n")
		w.Write(resp)
		return nil
	case InternalServerError:
		resp := []byte("HTTP/1.1 500 Internal Server Error\r\n")
		w.Write(resp)
		return nil
	default:
		resp := []byte("HTTP/1.1 201 \r\n")
		w.Write(resp)
		return nil
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	contentLenStr := strconv.Itoa(contentLen)
	headers["content-length"] = contentLenStr
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	headerResp := []byte{}
	for key, value := range headers {
		respStr := fmt.Sprintf("%s: %s\r\n", key, value)
		headerResp = append(headerResp, respStr...)
	}
	headerResp = append(headerResp, "\r\n"...)
	_, err := w.Write(headerResp)
	if err != nil {
		fmt.Println("Error in WriteHeaders.")
		return err
	}
	return nil
}
