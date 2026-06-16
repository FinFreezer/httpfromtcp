package headers

import (
	"errors"
	"slices"
	"strings"
	"unicode"
)

type Headers map[string]string

var SpecialChars = []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	const CRLF = "\r\n"
	if !strings.Contains(string(data), CRLF) {
		return 0, false, nil
	}
	if string(data[:2]) == CRLF {
		return 2, true, nil
	}
	headerString := string(data)
	splitIndx := strings.Index(headerString, CRLF)

	//Host: localhost:42069\r\n
	headerLine := headerString[:splitIndx]

	err = checkForWhitespace([]byte(headerLine))
	if err != nil {
		return 0, false, err
	}
	headerFields := strings.SplitN(headerLine, ":", 2)

	if len(headerFields) != 2 {
		return 0, false, errors.New("Malformed header.")
	}

	headerKey := strings.ToLower(strings.Trim(headerFields[0], " "))
	headerValue := strings.Trim(headerFields[1], " ")
	err = checkForForbiddenCharacters(headerKey)
	if err != nil {
		return 0, false, err
	}

	if _, ok := h[headerKey]; !ok {
		h[headerKey] = headerValue
	} else {
		h[headerKey] = h[headerKey] + ", " + headerValue
	}

	bytesUsed := splitIndx + 2
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

func checkForForbiddenCharacters(s string) error {
	if len(s) < 1 {
		return errors.New("Invalid character found.")
	}

	for _, char := range s {

		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			if ok := slices.Contains(SpecialChars, char); !ok {
				return errors.New("Invalid character found.")
			}
		}

	}
	return nil

}
