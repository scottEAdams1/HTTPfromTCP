package request

import (
	"HTTPfromTCP/internal/headers"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	requestStateInitialised requestState = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	request := Request{
		Headers: make(headers.Headers),
		state:   0,
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
				if request.state != requestStateDone {
					return nil, errors.New("Incomplete request")
				}
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
	totalBytesParsed := 0
	for r.state != requestStateDone {
		numBytes, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += numBytes
		if numBytes == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	if r.state == requestStateInitialised {
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		}
		r.RequestLine = requestLine
		r.state = requestStateParsingHeaders
		return numBytes, nil
	} else if r.state == requestStateParsingHeaders {
		numBytes, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done == true {
			r.state = requestStateParsingBody
		}
		return numBytes, err
	} else if r.state == requestStateParsingBody {
		if r.Headers.Get("Content-Length") == "" {
			r.state = requestStateDone
			return len(data), nil
		}
		r.Body = append(r.Body, data...)
		length, err := strconv.Atoi(r.Headers.Get("Content-Length"))
		if err != nil {
			return 0, err
		}
		if len(r.Body) > length {
			return 0, errors.New("Incorrect content length")
		} else if len(r.Body) == length {
			r.state = requestStateDone
		}
		return len(data), nil

	} else if r.state == requestStateDone {
		return 0, errors.New("Trying to read data in done state")
	} else {
		return 0, errors.New("Unknown state")
	}
}
