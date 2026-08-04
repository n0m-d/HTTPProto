package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"example.com/HProtocol/internal/request"
	"example.com/HProtocol/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) Write(w io.Writer) error {
	if err := response.WriteStatusLine(w, h.StatusCode); err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(h.Message))
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}
	_, err := w.Write([]byte(h.Message))
	return err
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		handler:  handler,
		listener: listener,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	body := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(len(body))
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}
