package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	var err error
	go func() {
		defer f.Close()
		defer close(out)

		line := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}
			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				line += string(data[:i])
				data = data[i+1:]
				out <- string(line)
				line = ""
			}
			line += string(data)
		}
		if err != nil && err != io.EOF {
			slog.Error("Error reading file", "error", err)
			close(out)
			return
		}
		if len(line) != 0 {
			out <- line
		}
	}()
	return out
}

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

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}
		log.Info("Connection closed after reading")
	}
}
