package request

import (
	"bytes"
	"errors"
	"io"

	"example.com/HProtocol/internal/headers"
)

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       parserState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

var (
	CRLF                         = "\r\n"
	MALFORMED_REQUEST            = errors.New("Malformed Request Header")
	UNSUPPORTED_HTTP_VERSION     = errors.New("Unsupported/ Malformed HTTP Version")
	ERROR_REQUEST_IN_ERROR_STATE = errors.New("Request in error state")
)

type parserState string

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func (r *Request) error() bool {
	return r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state:   StateInit,
		Headers: headers.Headers{},
	}
	buff := make([]byte, 1024)
	buffLen := 0

	for !request.done() {
		n, err := reader.Read(buff[buffLen:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		buffLen += n //Increases the total number of bytes in the buffer to include the newly read data.
		readN, err := request.Parse(buff[:buffLen])
		if err != nil {
			return nil, err
		}

		// Shift remaining bytes to the front
		copy(buff, buff[readN:buffLen])
		buffLen -= readN

		if err == io.EOF || n == 0 {
			break
		}
	}

	if !request.done() {
		return nil, MALFORMED_REQUEST
	}

	return request, nil
}

func (r *Request) Parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, nil
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateDone:
			break outer

		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n
			if done {
				r.state = StateDone
			}
		}
	}
	return read, nil
}

func parseRequestLine(reqs []byte) (*RequestLine, int, error) {
	idx := bytes.Index(reqs, []byte(CRLF))
	if idx == -1 {
		// No full line yet — wait for more data
		return nil, 0, nil
	}

	startLine := reqs[:idx]
	read := idx + len(CRLF)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, MALFORMED_REQUEST
	}

	versionParts := bytes.Split(parts[2], []byte("/"))
	if len(versionParts) != 2 || string(versionParts[0]) != "HTTP" {
		return nil, 0, UNSUPPORTED_HTTP_VERSION
	}

	version := string(versionParts[1])
	if version != "1.1" {
		return nil, 0, UNSUPPORTED_HTTP_VERSION
	}

	return &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   version,
	}, read, nil
}
