package request

import (
	"fmt"
	"http-from-tcp/internal/headers"
	"io"
	"slices"
	"strconv"
	"strings"
)

type parserState int

const (
	RequestStateInit parserState = iota
	RequestStateParsingHeader
	RequestStateParsingBody
	RequestStateParsed
)

const SEPARATOR = "\r\n"

var ValidMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

var ErrorMalformedStartLine = fmt.Errorf("bad request line")
var ErrorInvalidData = fmt.Errorf("invalid data")
var ErrorInvalidRequestLine = fmt.Errorf("invalid request line")
var ErrorParsingRequestLine = fmt.Errorf("unable to parse request line even after parsing the complete data sent")
var ErrorBodyLengthExceeded = fmt.Errorf("body length is more than the content-length header")
var ErrorReadingBody = fmt.Errorf("error reading the body")

type RequestLine struct {
	HttpVersion   string
	Method        string
	RequestTarget string
}

func (r *RequestLine) ValidateRequestLine() bool {
	if r.HttpVersion != "1.1" {
		return false
	}

	if !slices.Contains(ValidMethods, r.Method) {
		return false
	}

	return true
}

func getInt(val string, defValue int) (int, error) {
	if val == "" {
		return defValue, nil
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return intVal, nil
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       parserState
}

func NewRequest() *Request {
	return &Request{
		State:   RequestStateInit,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]

		switch r.State {
		case RequestStateInit:
			requestLine, parsedLength, err := parseRequestLine(string(data))
			if err != nil {
				return 0, err
			}
			if parsedLength == 0 {
				break outer
			}

			r.RequestLine = *requestLine

			r.State = RequestStateParsingHeader
			read = parsedLength

		case RequestStateParsingHeader:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}

			read += n

			if done {
				r.State = RequestStateParsingBody
			}

		case RequestStateParsingBody:
			contentLength, err := getInt(r.Headers.Get("Content-Length"), 0)
			if err != nil {
				return 0, err
			}

			if contentLength == 0 {
				r.State = RequestStateParsed
				break outer
			}

			if len(currentData) == 0 {
				if len(r.Body) != contentLength {
					return 0, ErrorReadingBody
				}
				r.State = RequestStateParsed
				break outer
			}

			if len(r.Body)+len(currentData) > contentLength {
				return 0, ErrorBodyLengthExceeded
			}

			r.Body = append(r.Body, currentData...)
			read += len(currentData)
			if len(r.Body) == contentLength {
				r.State = RequestStateParsed
				break outer
			}
			break outer

		case RequestStateParsed:
			break outer

		}
	}
	return read, nil
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

	return rl, len(before) + len(SEPARATOR), nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()
	buf := make([]byte, 1024)
	bufIdx := 0
	for request.State != RequestStateParsed {
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

		if err == io.EOF && request.State != RequestStateParsed {
			return nil, ErrorParsingRequestLine
		}
	}

	return request, nil
}
