package main

import (
	"HTTPfromTCP/internal/request"
	"HTTPfromTCP/internal/response"
	"HTTPfromTCP/internal/server"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
		return
	} else if path == "/video" {
		videoHandler(w, req)
		return
	} else if path == "/yourproblem" {
		handler400(w, req)
		return
	} else if path == "/myproblem" {
		handler500(w, req)
		return
	} else {
		handler200(w, req)
		return
	}
}

func handler200(w *response.Writer, req *request.Request) {
	err := w.WriteStatusLine(response.OK)
	if err != nil {
		fmt.Println("Error writing status line:", err.Error())
		return
	}
	message := []byte("<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>")
	handlerDefault(w, message)
}

func handler400(w *response.Writer, req *request.Request) {
	err := w.WriteStatusLine(response.BADREQUEST)
	if err != nil {
		fmt.Println("Error writing status line:", err.Error())
		return
	}
	message := []byte("<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>")
	handlerDefault(w, message)
}

func handler500(w *response.Writer, req *request.Request) {
	err := w.WriteStatusLine(response.SERVERERROR)
	if err != nil {
		fmt.Println("Error writing status line:", err.Error())
		return
	}
	message := []byte("<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>")
	handlerDefault(w, message)
}

func handlerDefault(w *response.Writer, message []byte) {
	headers := w.GetDefaultHeaders(len(message))
	err := w.WriteHeaders(headers)
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

func proxyHandler(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.OK)
	headers := w.GetDefaultHeaders(0)
	delete(headers, "Content-Length")
	headers["Transfer-Encoding"] = "chunked"
	headers["Trailer"] = "X-Content-SHA256,X-Content-Length"
	w.WriteHeaders(headers)
	res, err := http.Get("https://httpbin.org" + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin"))
	if err != nil {
		return
	}
	var full []byte
	for true {
		b := make([]byte, 1024)
		n, err := res.Body.Read(b)
		full = append(full, b[:n]...)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if n > 0 {
					w.WriteChunkedBody(b[:n])
				}
				hash := fmt.Sprintf("%x", sha256.Sum256(full))
				headers["X-Content-SHA256"] = hash
				headers["X-Content-Length"] = strconv.Itoa(len(full))
				w.WriteTrailers(headers)
				break
			}
			return
		}
		w.WriteChunkedBody(b[:n])
	}
}

func videoHandler(w *response.Writer, req *request.Request) {
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		handler500(w, req)
		return
	}
	w.WriteStatusLine(response.OK)
	headers := w.GetDefaultHeaders(len(data))
	headers["Content-Type"] = "video/mp4"
	w.WriteHeaders(headers)
	w.WriteBody(data)
}
