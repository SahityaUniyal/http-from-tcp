package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var ErrorMalformedHeader = fmt.Errorf("malformed header")
var ErrorMalformedHeaderName = fmt.Errorf("malformed header name")

var CRLF = []byte("\r\n")

type Headers map[string]string

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)

	if val, ok := h[key]; ok {
		value = fmt.Sprintf("%s, %s", val, value)
	}

	h[key] = value
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Delete(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func (h Headers) isToken(str string) bool {
	for _, ch := range str {
		found := false

		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			found = true
		}

		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	valueSeparator := []byte(":")
	// Split on the first value separator
	parts := bytes.SplitN(fieldLine, valueSeparator, 2)
	if len(parts) != 2 {
		return "", "", ErrorMalformedHeader
	}
	name := string(parts[0])
	value := strings.Trim(string(parts[1]), " ")

	// no trailing space allowed
	if strings.HasSuffix(name, " ") {
		return "", "", ErrorMalformedHeaderName
	}
	name = strings.Trim(name, " ")

	return name, value, nil
}
func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], CRLF)
		if idx == -1 {
			break
		}

		// Empty Headers
		if idx == 0 {
			done = true
			read += len(CRLF)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		read += idx + len(CRLF)

		if !h.isToken(name) {
			return 0, false, ErrorMalformedHeaderName
		}

		h.Set(name, value)
	}

	return read, done, nil
}
