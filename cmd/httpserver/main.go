package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcptohttp/internal/request"
	"tcptohttp/internal/response"
	"tcptohttp/internal/server"
)

const port = 42069

func _400BadRequest() []byte {
	return []byte(
		`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func _500Internal() []byte {
	return []byte(
		`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func _200Ok() []byte {
	return []byte(
		`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>>`)
}

func main() {
	s, err := server.Serve(port, func(w response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := _200Ok()
		status := response.StatusOK

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body = _400BadRequest()
			status = response.StatusBadRequest
		case "/myproblem":
			body = _500Internal()
			status = response.StatusInternalServerError
		}

		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")
		w.WriteStatusLine(status)
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	defer s.Close()

	log.Println("Server start on port: ", port)

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	log.Println("Server gracefully stopped")
}
