package headers

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	const CRLF = "\r\n"
	if !strings.Contains(string(data), CRLF) {
		return 0, false, nil
	}
	if string(data[:2]) == CRLF {
		return 2, true, nil
	}
	//Host: localhost:42069\r\n
	headerLine := strings.Split(string(data), "\r\n")
	fmt.Println(headerLine)

	err = checkForWhitespace([]byte(headerLine[0]))
	if err != nil {
		return 0, false, err
	}
	headerFields := strings.SplitN(headerLine[0], ":", 2)
	headerKey := strings.Trim(headerFields[0], " ")
	headerValue := strings.Trim(headerFields[1], " ")

	if _, ok := h[headerKey]; !ok {
		h[headerKey] = headerValue
	}

	bytesUsed := len(headerLine[0] + "\r\n")
	return bytesUsed, false, nil

}

func checkForWhitespace(d []byte) error {
	s := string(d)
	firstSpace := rune(s[0])
	if !unicode.IsLetter(firstSpace) {
		return errors.New("Faulty formatting.")
	}
	indx := strings.Index(s, ":")
	if indx == -1 {
		return errors.New("Faulty formatting.")
	}
	beforeColon := rune(s[indx-1])
	if unicode.IsSpace(beforeColon) {
		return errors.New("Faulty formatting.")
	}
	return nil
}
