package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/finfreezer/httpfromtcp/internal/headers"
)

type StatusCode int
type StatusMap map[StatusCode]string
type WriterStatusCode string

var (
	StatusMessages StatusMap = initializeStatus()
)

const (
	OK                  StatusCode       = 200
	BadRequest          StatusCode       = 400
	InternalServerError StatusCode       = 500
	AwaitStatus         WriterStatusCode = "AwaitStatusLine"
	AwaitHeaders        WriterStatusCode = "AwaitHeaders"
	AwaitBody           WriterStatusCode = "AwaitBody"
)

const respTemplate = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

type Writer struct {
	ContentWriter io.Writer
	ResponseHTML  string
	Headers       headers.Headers
	//WriterStatus WriterStatusCode
}

func NewResponseWriter() *Writer {
	newWriter := Writer{ResponseHTML: respTemplate, Headers: headers.NewHeaders()}
	return &newWriter
}

func initializeStatus() StatusMap {
	statusMap := make(StatusMap)
	statusMap[OK] = "OK"
	statusMap[BadRequest] = "Bad Request"
	statusMap[InternalServerError] = "Internal Server Error"
	return statusMap
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	replacement := ""
	/*if w.WriterStatus != AwaitStatus {
		return errors.New("Invalid status, Status Line potentially already written.")
	}*/

	startIndx := strings.Index(w.ResponseHTML, "<title>") + len("<title>")
	endIndx := strings.Index(w.ResponseHTML, "</title>")
	if startIndx == -1 || endIndx == -1 {
		return errors.New("HTML misformatting.")
	}
	statusCodeStr := strconv.Itoa(int(statusCode))

	if _, ok := StatusMessages[statusCode]; ok {
		replacement = fmt.Sprintf("%s %s", statusCodeStr, StatusMessages[statusCode])
	} else {
		replacement = fmt.Sprintf("%s", statusCodeStr)
	}

	w.ResponseHTML = w.ResponseHTML[:startIndx] + replacement + w.ResponseHTML[endIndx:]
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	contentLenStr := strconv.Itoa(contentLen)
	headers["content-length"] = contentLenStr
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func SetDefaultHeaders(headers headers.Headers, KeysToChange, Values []string) headers.Headers {
	for i := range KeysToChange {
		headers[KeysToChange[i]] = Values[i]
	}
	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	/*if w.WriterStatus != AwaitHeaders {
		return errors.New("Invalid status, headers potentially already written, or awaiting status line.")
	}*/

	headerResp := []byte{}
	for key, value := range headers {
		respStr := fmt.Sprintf("%s: %s\r\n", key, value)
		headerResp = append(headerResp, respStr...)
	}
	headerResp = append(headerResp, "\r\n"...)
	_, err := w.ContentWriter.Write(headerResp)
	if err != nil {
		fmt.Println("Error in WriteHeaders.")
		return err
	}
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	/*if w.WriterStatus != AwaitBody {
		return 0, errors.New("Invalid status, headers potentially already written, or awaiting status line.")
	}*/
	startIndx := strings.Index(w.ResponseHTML, "<p>") + len("<p>")
	endIndx := strings.Index(w.ResponseHTML, "</p>")
	if startIndx == -1 || endIndx == -1 {
		return 0, errors.New("HTML misformatting.")
	}

	w.ResponseHTML = w.ResponseHTML[:startIndx] + string(p) + w.ResponseHTML[endIndx:]
	return len(p), nil
}

func (w *Writer) ReplaceHTMLHeader(p []byte) error {
	startIndx := strings.Index(w.ResponseHTML, "<h1>") + len("<h1>")
	endIndx := strings.Index(w.ResponseHTML, "</h1>")
	if startIndx == -1 || endIndx == -1 {
		return errors.New("HTML misformatting.")
	}

	w.ResponseHTML = w.ResponseHTML[:startIndx] + string(p) + w.ResponseHTML[endIndx:]
	return nil
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
