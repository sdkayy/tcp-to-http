package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func getLinesChannel(conn net.Conn) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer conn.Close()

		strs := ""
		for {
			buffer := make([]byte, 8)
			b, err := conn.Read(buffer)
			if err != nil && err != io.EOF {
				panic(err)
			}

			if err == io.EOF {
				if strs != "" {
					ch <- strs
				}
				break
			}

			if b == 0 {
				continue
			}

			strs += string(buffer[:b])
			for strings.Contains(strs, "\n") {
				parts := strings.SplitN(strs, "\n", 2)
				line := parts[0]
				ch <- line + "\n"
				if len(parts) == 2 {
					strs = parts[1]
				} else {
					strs = ""
				}
			}
		}

	}()

	return ch
}

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

		fmt.Println("Connection accepted")

		ch := getLinesChannel(conn)

		for str := range ch {
			fmt.Println(str)
		}

		fmt.Println("Connection Closed")
	}
}
