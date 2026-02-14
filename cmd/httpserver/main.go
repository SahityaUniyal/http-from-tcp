package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"http-from-tcp/internal/headers"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"http-from-tcp/internal/server"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

var resp400 = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

var resp500 = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

var resp200 = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)

func handler(w response.Writer, req *request.Request) {
	var body []byte
	var sc response.StatusCode

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		reqUrl := url.URL{
			Scheme: "https",
			Host:   "httpbin.org",
			Path:   strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin"),
		}
		resp, err := http.Get(reqUrl.String())
		if err != nil {
			he := server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    fmt.Sprintf("error making request to httpbin, %v", err),
			}
			slog.Info("reached")
			he.Write(w)
			return
		}
		// TODO: add a check for the response code
		w.WriteStatusLine(response.StatusOK)

		h := response.GetDefaultHeaders(0)
		h.Delete("content-length")
		h.Set("transfer-encoding", "chunked")
		h.Set("Trailer", "X-Content-SHA256")
		h.Set("Trailer", "X-Content-Length")
		h.Delete("content-length")
		w.WriteHeaders(h)

		respBody := make([]byte, 1024)
		fullBody := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(respBody)
			slog.Info("read data from httpbin", "data", n)
			if err != nil && err != io.EOF {
				he := server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    fmt.Sprintf("error reading body from httpbin, %v", err),
				}
				he.Write(w)
				return
			}
			if n == 0 || err == io.EOF {
				w.WriteChunkedBodyDone()
				break
			}
			fullBody = append(fullBody, respBody[:n]...)
			w.WriteChunkedBody(respBody[:n])
		}
		sum := sha256.Sum256(fullBody)
		hexSum := hex.EncodeToString(sum[:])
		trailer := headers.NewHeaders()
		trailer.Set("X-Content-SHA256", hexSum)
		trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
		err = w.WriteTrailers(trailer)
		w.Write([]byte("\r\n"))
		return
	}

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		body = resp400
		sc = response.StatusBadRequest
	case "/myproblem":
		body = resp500
		sc = response.StatusInternalServerError
	default:
		body = resp200
		sc = response.StatusOK
	}

	w.WriteStatusLine(sc)

	h := response.GetDefaultHeaders(0)
	h.Override("content-length", fmt.Sprintf("%d", len(body)))
	h.Override("content-type", "text/html")
	w.WriteHeaders(h)

	w.WriteBody(body)
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")

}
