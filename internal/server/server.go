package server

import (
	"HTTPfromTCP/internal/request"
	"HTTPfromTCP/internal/response"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	isClosed atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return &Server{}, err
	}
	server := &Server{
		Listener: listener,
		handler:  handler,
	}
	go func() {
		server.listen()
	}()
	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	err := s.Listener.Close()
	return err
}

func (s *Server) listen() {
	for {
		// Wait for a connection.
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			log.Println(err.Error())
			continue
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			s.handle(c)
			// Shut down the connection.
			c.Close()
		}(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	req, err := request.RequestFromReader(conn)
	var buf bytes.Buffer
	writer2 := response.Writer{
		Writer: &buf,
	}
	if err != nil {
		writer2.WriteStatusLine(response.BADREQUEST)
		body := "<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>" + err.Error() + "</p></body></html>"
		writer2.WriteHeaders(writer2.GetDefaultHeaders(len(body)))
		writer2.WriteBody([]byte(body))
		return
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(&writer2, req)
	} else {
		s.handler(&writer2, req)
	}
	conn.Write(buf.Bytes())
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

type Handler func(w *response.Writer, req *request.Request)
