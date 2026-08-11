package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"example.com/HProtocol/internal/headers"
	"example.com/HProtocol/internal/request"
	"example.com/HProtocol/internal/response"
	"example.com/HProtocol/internal/server"
)

const port = 42069

func toHTTPBingoURL(path string) string {
	return "https://httpbingo.org" + strings.TrimPrefix(path, "/httpbin")
}

func proxyHandler(w *response.Writer, req *request.Request) {
	upstreamURL := toHTTPBingoURL(req.RequestLine.RequestTarget)
	resp, err := http.Get(upstreamURL)
	if err != nil {
		w.WriteStatusLine(response.StatusInternalServerError)
		body := []byte(err.Error())
		h := response.GetDefaultHeaders(len(body))
		w.WriteHeaders(h)
		w.WriteBody(body)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(0)
	h.Delete("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	var fullBody []byte
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		fmt.Println("Read:", n)
		if n > 0 {
			fullBody = append(fullBody, buf[:n]...)
			_, writeErr := w.WriteChunkedBody(buf[:n])
			if writeErr != nil {
				return
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
	}

	hash := sha256.Sum256(fullBody)
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	w.WriteTrailers(trailers)
}

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			proxyHandler(w, req)
			return
		}

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`)
			w.WriteStatusLine(response.StatusBadRequest)
			h := response.GetDefaultHeaders(len(body))
			h.Replace("Content-Type", "text/html")
			w.WriteHeaders(h)
			w.WriteBody(body)
		case "/myproblem":
			body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`)
			w.WriteStatusLine(response.StatusInternalServerError)
			h := response.GetDefaultHeaders(len(body))
			h.Replace("Content-Type", "text/html")
			w.WriteHeaders(h)
			w.WriteBody(body)

		case "/video":
			body, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				h := response.GetDefaultHeaders(len(body))
				h.Replace("Content-Type", "video/mp4")
				w.WriteHeaders(h)
				w.WriteBody(body)
				return
			}
			w.WriteStatusLine(response.StatusOK)
			h := response.GetDefaultHeaders(len(body))
			h.Replace("Content-Type", "video/mp4")
			w.WriteHeaders(h)
			w.WriteBody(body)
		default:
			body := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`)
			w.WriteStatusLine(response.StatusOK)
			h := response.GetDefaultHeaders(len(body))
			h.Replace("Content-Type", "text/html")
			w.WriteHeaders(h)
			w.WriteBody(body)
		}
	}

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
