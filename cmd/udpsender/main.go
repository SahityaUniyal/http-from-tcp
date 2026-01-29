package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Error("error listening on port :42069", "error", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Error("error dialing UDP connection on port:42069", "error", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		data, err := reader.ReadBytes('\n')
		if err != nil {
			log.Error("error reading from STDIN", "error", err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			log.Error("error reading from STDIN", "error", err)
			return
		}
	}

}
