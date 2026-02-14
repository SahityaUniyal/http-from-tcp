package main

import (
	"fmt"
	"http-from-tcp/internal/request"
	"log/slog"
	"net"
	"os"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	lr, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Error("error listening on port :42069", "error", err)
		return
	}
	defer lr.Close()

	log.Info("Listening on port :42069")

	for {
		conn, err := lr.Accept()
		if err != nil {
			log.Error("error accepting the connection", "error", err)
			return
		}

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Error(fmt.Sprintf("error reading from the connection :: %v", err))
			return
		}
		fmt.Println("Request Line:")
		fmt.Printf("Method: %v\n", request.RequestLine.Method)
		fmt.Printf("Target: %v\n", request.RequestLine.RequestTarget)
		fmt.Printf("Version: %v\n", request.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %v: %v\n", key, value)
		}

		fmt.Println("Body:")
		fmt.Printf("%s\n", request.Body)
		log.Info("Connection closed after reading")
	}
}
