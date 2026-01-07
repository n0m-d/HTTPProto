package headers

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
)

type Headers map[string]string

var CRLF = []byte("\r\n")

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key string, value string) {
	key = strings.ToLower(key)
	if current, exists := h[key]; exists && current != "" {
		h[key] = current + ", " + value
	} else {
		h[key] = value
	}
}

func isvalidFieldName(str []byte) bool {
	if len(str) == 0 {
		return false
	}
	if bytes.ContainsAny(str, " \t") {
		return false
	}
	pattern := `^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(string(str))
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

		if !isvalidFieldName([]byte(key)) {
			return 0, false, errors.New("malformed  header name")
		}

		h.Set(key, value)

	}

	return read, done, nil
}
