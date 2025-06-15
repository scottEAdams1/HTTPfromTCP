package server

import (
	"HTTPfromTCP/internal/request"
	"HTTPfromTCP/internal/response"
	"bytes"
	"log"
	"net"
	"strconv"
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
	s.handler(&writer2, req)
	conn.Write(buf.Bytes())
}

type Handler func(w *response.Writer, req *request.Request)
