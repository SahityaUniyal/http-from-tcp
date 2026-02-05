package request

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
)

type parserState int

const (
	StateInit parserState = iota
	StateParsed
)

const SEPARATOR = "\r\n"

var ValidMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

var ErrorMalformedStartLine = fmt.Errorf("bad request line")
var ErrorInvalidData = fmt.Errorf("invalid data")
var ErrorInvalidRequestLine = fmt.Errorf("invalid request line")
var ErrorParsingRequestLine = fmt.Errorf("unable to parse request line even after parsing the complete data sent")

type RequestLine struct {
	HttpVersion   string
	Method        string
	RequestTarget string
}

func (r *RequestLine) ValidateRequestLine() bool {
	if r.HttpVersion != "1.1" {
		slog.Error("unsupported http version")
		return false
	}

	if !slices.Contains(ValidMethods, r.Method) {
		slog.Error("unsupported http method")
		return false
	}

	return true
}

type Request struct {
	RequestLine RequestLine
	// Headers     map[string]string
	// Body        []byte
	State parserState
}

func NewRequest() *Request {
	return &Request{
		State: StateInit,
	}
}

func (r *Request) parse(data []byte) (int, error) {
	requestLine, parsedLength, err := parseRequestLine(string(data))
	if err != nil {
		return 0, err
	}
	if parsedLength == 0 {
		return 0, nil
	}

	r.RequestLine = *requestLine

	r.State = StateParsed

	return parsedLength, nil
}

func parseRequestLine(data string) (*RequestLine, int, error) {
	before, _, ok := strings.Cut(data, SEPARATOR)
	if !ok {
		return nil, 0, nil
	}
	startLine := before

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return nil, -1, ErrorMalformedStartLine
	}
	version := strings.Split(parts[2], "/")
	if len(version) != 2 {
		return nil, -1, ErrorMalformedStartLine
	}

	rl := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version[1],
	}
	if !rl.ValidateRequestLine() {
		return nil, -1, ErrorInvalidRequestLine
	}

	return rl, len(before), nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()
	buf := make([]byte, 1024)
	bufIdx := 0
	for request.State != StateParsed {
		if bufIdx == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[bufIdx:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		bufIdx += n
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}

		// Removing the parsed data from the buffer
		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN

		if err == io.EOF && request.State != StateParsed {
			return nil, ErrorParsingRequestLine
		}
	}

	return request, nil

}
