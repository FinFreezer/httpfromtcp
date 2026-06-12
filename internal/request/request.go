package request

import (
	"errors"
	"io"
	"log"
	"strings"
	"unicode"
)

var bufferSize = 8

type Request struct {
	RequestLine RequestLine
	Status      int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	newRqst := Request{Status: 0}

	for {
		if newRqst.Status != 1 {

			if readToIndex == len(buf) {
				newBuf := make([]byte, len(buf)*2)
				copy(newBuf, buf)
				buf = newBuf
			}

			n, err := reader.Read(buf[readToIndex:])
			if err == io.EOF {
				newRqst.Status = 1
			}
			readToIndex += n
			n, err = newRqst.parse(buf[:readToIndex])
			if err != nil {
				log.Printf("Error parsing buffer: %s", err)
				return nil, err
			}
			copy(buf, buf[n:readToIndex])
			readToIndex -= n
			if newRqst.Status == 1 {
				break
			}
		}
	}
	return &newRqst, nil
}

func parseRequestLine(request []byte) (int, *RequestLine, error) {
	isFullLine := strings.Contains(string(request), "\r\n")
	if !isFullLine {
		return 0, nil, nil
	}
	requestLine := strings.Split(string(request), "\r\n")[0]
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

	return len(request), &newRequestLine, nil
}

func (r *Request) parse(data []byte) (int, error) {
	bytesParsed := 0
	if r.Status == 0 {
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
			r.Status = 1
			return bytesParsed, nil
		}
	}

	if r.Status == 1 {
		log.Println("File parsed.")
		return bytesParsed, errors.New("Error: trying to read data in done state.")
	}

	return bytesParsed, errors.New("Unknown state.")
}
