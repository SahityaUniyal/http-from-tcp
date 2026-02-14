package server

import (
	"fmt"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
)

type Handler func(w response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) Write(w io.Writer) {
	resp := response.NewWrite(w)
	err := resp.WriteStatusLine(h.StatusCode)
	if err != nil {
		slog.Error("error writing status line", "error", err)
		return
	}
	err = resp.WriteHeaders(response.GetDefaultHeaders(len(h.Message)))
	if err != nil {
		slog.Error("error writing header", "error", err)
		return
	}
	_, err = resp.WriteBody([]byte(h.Message))
	if err != nil {
		slog.Error("error writing body", "error", err)
		return
	}
}

type Server struct {
	running  atomic.Bool
	listener net.Listener
	handler  Handler
}

func NewServer(l net.Listener, h Handler) *Server {
	s := &Server{
		handler:  h,
		listener: l,
	}
	return s
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		statusCode := response.StatusInternalServerError

		if err == request.ErrorMalformedStartLine || err == request.ErrorInvalidData || err == request.ErrorInvalidRequestLine || err == request.ErrorBodyLengthExceeded {
			statusCode = response.StatusBadRequest
		}

		he := &HandlerError{
			StatusCode: statusCode,
			Message:    err.Error(),
		}
		he.Write(conn)
		return
	}

	respWriter := response.NewWrite(conn)
	s.handler(respWriter, req)
}

func (s *Server) listen() {
	s.running.Swap(true)
	for {
		conn, err := s.listener.Accept()
		if !s.running.Load() {
			return
		}
		if err != nil {
			slog.Error("error accepting connection", "error", err)
			return
		}
		go func() {
			slog.Info("connection recieved, handling the connection")
			s.handle(conn)
		}()
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := NewServer(listener, handler)
	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.running.Swap(false)
	return s.listener.Close()
}
