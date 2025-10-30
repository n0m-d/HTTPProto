package headers

import (
	"bytes"
	"errors"
)

type Headers map[string]string

var CRLF = []byte("\r\n")

func NewHeaders() Headers {
	return Headers{}
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(parts) != 2 {
		return "", "", errors.New("malformed  header")
	}
	key := parts[0]
	value := bytes.TrimSpace(parts[1])
	if bytes.HasSuffix(key, []byte(" ")) {
		return "", "", errors.New("malformed  header")
	}
	return string(key), string(value), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], CRLF)
		if idx == -1 {
			break
		}
		if idx == 0 {
			done = true
			read += len(CRLF)
			break
		}
		key, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return read, false, err
		}
		read += idx + len(CRLF)
		h[key] = value

	}

	return read, done, nil
}
