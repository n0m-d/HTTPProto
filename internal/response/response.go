package response

import (
	"fmt"
	"io"

	"example.com/HProtocol/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateChunkedBody
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  writerStateStatusLine, //First write the status line
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.state)
	}
	defer func() { w.state = writerStateHeaders }()
	return WriteStatusLine(w.writer, statusCode)
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.state)
	}
	defer func() { w.state = writerStateBody }()
	return WriteHeaders(w.writer, headers)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.state)
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writerStateBody && w.state != writerStateChunkedBody {
		return 0, fmt.Errorf("cannot write chunked body in state %d", w.state)
	}
	w.state = writerStateChunkedBody

	_, err := fmt.Fprintf(w.writer, "%x\r\n", len(p))
	if err != nil {
		return 0, err
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}
	_, err = w.writer.Write([]byte("\r\n"))
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writerStateBody && w.state != writerStateChunkedBody {
		return 0, fmt.Errorf("cannot write chunked body done in state %d", w.state)
	}
	n, err := w.writer.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return n, err
	}
	w.state = writerStateBody
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writerStateBody && w.state != writerStateChunkedBody {
		return fmt.Errorf("cannot write trailers in state %d", w.state)
	}
	_, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return err
	}
	err = WriteHeaders(w.writer, h)
	if err != nil {
		return err
	}
	w.state = writerStateBody
	return nil
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := ""
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	}
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	return err
}
