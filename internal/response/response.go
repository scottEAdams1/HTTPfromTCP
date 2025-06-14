package response

import (
	"HTTPfromTCP/internal/headers"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

type StatusCode int

const (
	OK          StatusCode = 200
	BADREQUEST  StatusCode = 400
	SERVERERROR StatusCode = 500
)

type Writer struct {
	Writer io.Writer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if statusCode == OK {
		w.Writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return nil
	}
	if statusCode == BADREQUEST {
		w.Writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return nil
	}
	if statusCode == SERVERERROR {
		w.Writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return nil
	}
	return errors.New("Incorrect status code")
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for k, v := range headers {
		_, err := w.Writer.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.Writer.Write(p)
}

func (w *Writer) GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/html"
	return headers
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	length := len(p)
	hex := fmt.Sprintf("%x", length)
	n, err := w.Writer.Write([]byte(hex + "\r\n"))
	if err != nil {
		return n, err
	}
	n2, err := w.Writer.Write(p)
	if err != nil {
		return n + n2, err
	}
	n3, err := w.Writer.Write([]byte("\r\n"))
	return n + n2 + n3, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.Writer.Write([]byte("0\r\n\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	_, err := w.Writer.Write([]byte("0\r\n"))
	if err != nil {
		return err
	}
	trailers := strings.Split(h["Trailer"], ",")
	for k, v := range h {
		if slices.Contains(trailers, k) {
			trailer := fmt.Sprintf("%s: %s\r\n", k, v)
			_, err = w.Writer.Write([]byte(trailer))
			if err != nil {
				return err
			}
		}
	}
	_, err = w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
