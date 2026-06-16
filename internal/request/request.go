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
	isEOF := false

	for newRqst.State != Finished {

		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			isEOF = true
		}
		if err != nil && !isEOF {
			return nil, err
		}
		readToIndex += n

		for readToIndex > 0 {
			log.Printf("state before parse: %v, buffered=%d", newRqst.State, readToIndex)
			n, err = newRqst.parse(buf[:readToIndex])
			log.Printf("state after parse: %v, consumed=%d, err=%v", newRqst.State, n, err)
			if err != nil {
				return nil, err
			}
			if n > 0 {
				copy(buf, buf[n:readToIndex])
				readToIndex -= n
			}
			if newRqst.State == Finished {
				break
			}
			if n == 0 {
				break
			}
		}

		if newRqst.State == Finished {
			break
		}
		if isEOF {
			return nil, errors.New("Reached EOF without finishing request.")
		}
	}
	return &newRqst, nil
}

func parseRequestLine(request []byte) (int, *RequestLine, error) {
	requestLine := string(request)

	splitPos := strings.Index(requestLine, "\r\n")
	if splitPos == -1 {
		return 0, nil, nil
	}
	linePart := requestLine[:splitPos]
	requestLineParts := strings.Split(linePart, " ")

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
	bytesConsumed := splitPos + 2
	return bytesConsumed, &newRequestLine, nil
}

func (r *Request) parse(data []byte) (int, error) {
	bytesParsed := 0
	log.Printf("Current state: %v\n", r.State)

	if r.State == ParseBody {
		length := r.Headers.Get("content-length")
		log.Printf("Parsing body, expecting %s bytes, remaining data %d\n", length, len(data))
		if length == "" {
			log.Println("No content-length key.")
			r.State = Finished
			return 0, nil
		}
		lengthInt, err := strconv.Atoi(length)
		if err != nil {
			return 0, err
		}
		bytesRequired := lengthInt - len(r.Body)
		bytesParsed := min(bytesRequired, len(data))
		r.Body = append(r.Body, data[:bytesParsed]...)

		if len(r.Body) == lengthInt {
			r.State = Finished
			log.Printf("Body Parsed... new state: %+v\n", r.State)
			return bytesParsed, nil
		}

		return bytesParsed, nil
	}

	if r.State == ParseHeaders {
		n, doneState, err := r.Headers.Parse(data[bytesParsed:])
		if err != nil {
			return 0, err
		}
		bytesParsed += n
		if doneState {
			r.State = ParseBody
			contentLength := r.Headers.Get("content-length")
			if contentLength == strconv.Itoa(0) || contentLength == "" {
				r.State = Finished
			}
			log.Printf("Headers Parsed... new state: %+v\n", r.State)
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
			log.Printf("Request Line Parsed... new state: %+v", r.State)
			return bytesParsed, nil
		}
	}

	if r.State == Finished {
		log.Println("Parsing finished.")
		return bytesParsed, nil
	}
	return bytesParsed, errors.New("Unknown state.")
}
