package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
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
	fs, err := os.Open("message.txt")
	if err != nil {
		log.Error("error opening file", "error", err.Error())
	}
	lines := getLinesChannel(fs)
	for line := range lines {
		fmt.Printf("%s\n", line)
	}
}
