package request

import (
	"errors"
	//"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	h "github.com/finfreezer/httpfromtcp/internal/headers"
)

var bufferSize = 8

type RequestStatus int

const (
	ParseRqstLine RequestStatus = iota
	ParseHeaders
	ParseBody
	Finished
)

type Request struct {
	RequestLine RequestLine
	Headers     h.Headers
	Body        []byte
	State       RequestStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	newRqst := Request{State: ParseRqstLine, Headers: h.NewHeaders()}

	for {
		if newRqst.State != Finished {

			if readToIndex == len(buf) {
				newBuf := make([]byte, len(buf)*2)
				copy(newBuf, buf)
				buf = newBuf
			}

			n, err := reader.Read(buf[readToIndex:])
			if err == io.EOF {
				newRqst.State = Finished
			}
			readToIndex += n
			n, err = newRqst.parse(buf[:readToIndex])
			if err != nil {
				log.Printf("Error parsing buffer: %s", err)
				if err.Error() != "Error: trying to read data in done state." {
					return nil, err
				}
				newRqst.State = Finished
			}
			copy(buf, buf[n:readToIndex])
			readToIndex -= n
			if newRqst.State == Finished {
				break
			}
		}
	}
	//helperPrintRequest(newRqst)
	return &newRqst, nil
}

func parseRequestLine(request []byte) (int, *RequestLine, error) {
	isFullLine := strings.Contains(string(request), "\r\n")
	if !isFullLine {
		return 0, nil, nil
	}
	requestLine := strings.Split(string(request), "\r\n")[0]
	leftOvers := strings.Split(string(request), "\r\n")[1]
	requestLineParts := strings.Split(requestLine, " ")

	if len(requestLineParts) != 3 {
		log.Println("Incorrect amount of parts in request.")
		return 0, nil, errors.New("Wrong amount of parts.")
	}

	HTTPVersionFull := strings.Split(requestLineParts[2], "/")
	HTTPVersion := HTTPVersionFull[1]
	Method := requestLineParts[0]
	RequestTarget := requestLineParts[1]

	newRequestLine := RequestLine{
		HttpVersion:   HTTPVersion,
		RequestTarget: RequestTarget,
		Method:        Method,
	}

	for _, char := range newRequestLine.Method {
		if !unicode.IsLetter(char) {
			log.Printf("Incorrect request method: %s", newRequestLine.Method)
			return len(request), nil, errors.New("Incorrect request method")
		}
	}
	if newRequestLine.HttpVersion != "1.1" {
		log.Printf("Incorrect HTTP version.")
		return len(request), nil, errors.New("Incorrect HTTP version")
	}

	return len(request) - len(leftOvers), &newRequestLine, nil
}

func (r *Request) parse(data []byte) (int, error) {
	bytesParsed := 0

	if r.State == ParseBody {
		length := r.Headers.Get("content-length")
		if length == "" {
			r.State = Finished
			return 0, nil
		}
		lengthInt, err := strconv.Atoi(length)
		if err != nil {
			return 0, err
		}
		log.Println(string(data))
		r.Body = append(r.Body, data...)

		if len(r.Body) != lengthInt {
			return 0, errors.New("Body length doesn't match content-length.")
		}
		r.State = Finished
		return len(data), nil
	}

	if r.State == ParseHeaders {
		n, doneState, err := r.Headers.Parse(data[bytesParsed:])
		if err != nil {
			return 0, err
		}
		bytesParsed += n
		if doneState {
			r.State = ParseBody
		}
		return bytesParsed, nil
	}

	if r.State == ParseRqstLine {
		bytesParsed, req, err := parseRequestLine(data)
		if err != nil {
			log.Printf("Error parsing content: %s", err)
			return 0, err
		}
		if bytesParsed == 0 {
			return 0, nil
		}
		if req != nil {
			r.RequestLine = *req
			r.State = ParseHeaders
			return bytesParsed, nil
		}
	}

	if r.State == Finished {
		log.Println("Parsing finished.")
		return bytesParsed, errors.New("Error: trying to read data in done state.")
	}
	return bytesParsed, errors.New("Unknown state.")
}

/*func helperPrintRequest(r Request) {
	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for key, value := range r.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
}*/
