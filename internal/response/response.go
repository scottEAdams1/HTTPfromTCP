package response

import (
	"HTTPfromTCP/internal/headers"
	"errors"
	"io"
	"strconv"
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
