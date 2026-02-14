package response

import (
	"errors"
	"fmt"
	"http-from-tcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode uint16

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writeState uint16

const (
	NothingWritten writeState = iota
	StatusLineWritten
	HeaderWritten
	BodyWritten
)

const HTTPVersion = "1.1"

func getReason(sc StatusCode) string {
	switch sc {
	case StatusOK:
		return "OK"
	case StatusBadRequest:
		return "Bad Request"
	case StatusInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	len := strconv.Itoa(contentLen)

	h.Set("Content-Length", len)
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

type Writer struct {
	io.Writer
	state writeState
}

func NewWrite(w io.Writer) Writer {
	return Writer{
		Writer: w,
		state:  NothingWritten,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != NothingWritten {
		return errors.New("incorrect order should be written first")
	}
	statusLine := fmt.Sprintf("HTTP/%s %v %s\r\n", HTTPVersion, statusCode, getReason(statusCode))
	_, err := w.Write([]byte(statusLine))

	w.state = StatusLineWritten
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != StatusLineWritten {
		return errors.New("incorrect order should be written after status line")
	}
	for key, value := range headers {
		header := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(header))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.state = HeaderWritten
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.state != HeaderWritten {
		return 0, errors.New("incorrect order should be written after header")
	}

	w.state = BodyWritten
	return w.Write(body)
}
