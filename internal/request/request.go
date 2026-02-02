package request

import (
	"errors"
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

var VALID_METHODS = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

var ERROR_MALFORMED_START_LINE = fmt.Errorf("bad request line")
var ERROR_INVALID_DATA = fmt.Errorf("invalid data")
var ERROR_INVALID_REQUESTLINE = fmt.Errorf("invalid request line")

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

	if !slices.Contains(VALID_METHODS, r.Method) {
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
	requestLine, parsedLength, _, err := parseRequestLine(string(data))
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

func parseRequestLine(data string) (*RequestLine, int, string, error) {
	before, after, ok := strings.Cut(data, SEPARATOR)
	if !ok {
		return nil, 0, data, nil
	}
	startLine := before
	restOfMsg := after

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return nil, -1, "", ERROR_MALFORMED_START_LINE
	}
	version := strings.Split(parts[2], "/")
	if len(version) != 2 {
		return nil, -1, "", ERROR_MALFORMED_START_LINE
	}

	rl := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version[1],
	}
	if !rl.ValidateRequestLine() {
		return nil, -1, "", ERROR_INVALID_REQUESTLINE
	}

	return rl, len(before), restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()
	buf := make([]byte, 1024)
	bufIdx := 0
	for request.State != StateParsed {
		n, err := reader.Read(buf[bufIdx:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		bufIdx += n
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}
		if readN == 0 {
			continue
		}

		if err == io.EOF && request.State != StateParsed {
			return nil, errors.New("unable to parse request line even after parsing the complete data sent")
		}
	}

	return request, nil

}
