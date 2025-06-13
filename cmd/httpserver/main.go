package main

import (
	"HTTPfromTCP/internal/headers"
	"HTTPfromTCP/internal/request"
	"HTTPfromTCP/internal/response"
	"HTTPfromTCP/internal/server"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, serverHandler)
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

func serverHandler(w *response.Writer, req *request.Request) {
	path := req.RequestLine.RequestTarget
	var statusCode response.StatusCode
	var headers headers.Headers
	var message []byte
	if path == "/yourproblem" {
		statusCode = response.BADREQUEST
		message = []byte("<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>")
	} else if path == "/myproblem" {
		statusCode = response.SERVERERROR
		message = []byte("<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>")
	} else {
		statusCode = response.OK
		message = []byte("<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>")
	}
	headers = w.GetDefaultHeaders(len(message))
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		fmt.Println("Error writing status line:", err.Error())
		return
	}
	err = w.WriteHeaders(headers)
	if err != nil {
		fmt.Println("Error writing headers:", err.Error())
		return
	}
	n, err := w.WriteBody(message)
	if err != nil {
		fmt.Printf("Error writing body after %v bytes: %s", n, err.Error())
		return
	}
}
