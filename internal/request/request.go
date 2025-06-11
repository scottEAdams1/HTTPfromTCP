package request

import (
	"errors"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	request := Request{
		state: 0,
	}
	for request.state != 1 {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, 2*len(buf))
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = 1
				break
			}
			return nil, err
		}
		readToIndex += n
		n, err = request.parse(buf[:readToIndex])
		if err != nil {
			return &Request{}, err
		}
		copy(buf, buf[n:])
		readToIndex -= n
	}
	return &request, nil
}

func parseRequestLine(request []byte) (RequestLine, int, error) {
	s := string(request)
	contains := strings.Contains(s, "\r\n")
	if contains == false {
		return RequestLine{}, 0, nil
	}
	parts := strings.Split(s, "\r\n")
	requestLine := parts[0]
	requestLineParts := strings.Split(requestLine, " ")
	if len(requestLineParts) != 3 {
		return RequestLine{}, 0, errors.New("Request line has incorrect number of parts")
	}
	method := requestLineParts[0]
	requestTarget := requestLineParts[1]
	httpVersion := requestLineParts[2]
	if regexp.MustCompile(`^[a-zA-Z]*$`).MatchString(method) == false {
		return RequestLine{}, 0, errors.New("Method must only contain alphabetic characters")
	}
	if httpVersion != "HTTP/1.1" {
		return RequestLine{}, 0, errors.New("HTTP version must be HTTP/1.1")
	}
	return RequestLine{
		HttpVersion:   "1.1",
		RequestTarget: requestTarget,
		Method:        method,
	}, len(requestLine) + len("\r\n"), nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == 0 {
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		}
		r.RequestLine = requestLine
		r.state = 1
		return numBytes, nil
	}
	if r.state == 1 {
		return 0, errors.New("Trying to read data in a done state")
	}
	return 0, errors.New("Unknown state")
}
