package main

import (
	"fmt"
	"log"
	"net"
	"tcptohttp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		fmt.Println("Error starting listener")
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection")
			break
		}

		r, err := request.RequestFromReader(conn)

		if err != nil {
			log.Fatal("error", "error", err)
		}

		fmt.Printf("Request Line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Printf("Headers:\n")
		r.Headers.ForEach(func(name, value string) {
			fmt.Printf("- %s: %s\n", name, value)
		})
	}
}
