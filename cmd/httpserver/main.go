package main

import (
	"fmt"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"http-from-tcp/internal/server"
	"log"
	"os"
	"os/signal"
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
