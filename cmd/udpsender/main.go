package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)


func main() {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:42069")
	if err != nil {
		e := fmt.Errorf("unable to resolve addr")
		fmt.Println(e.Error())
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		e := fmt.Errorf("unable to open connection: %v", err)
		fmt.Println(e.Error())
		return
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
            if err == io.EOF {
                break // End of file reached
            }
            fmt.Printf("error reading line: %v", err) // Log any other error
            break
        }
		_, e := conn.Write([]byte(line))
		if e != nil {
			fmt.Printf("error sending")
		}
	}
}