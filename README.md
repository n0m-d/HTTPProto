# HProtocol

An HTTP/1.1 server built from scratch in Go. No `net/http` server. Just TCP, request parsing, and writing responses by hand.

This started as a learning project (boot.dev style): parse HTTP over raw sockets, then grow it into a small server with handlers, chunked encoding, and a proxy.

## What's in here

```
cmd/
  httpserver/     # main HTTP server
  tcplistener/    # early TCP + request parsing playground
  udpsender/      # tiny UDP sender for earlier exercises
internal/
  headers/        # header parse + store
  request/        # request line, headers, body
  response/       # status line, headers, body, chunks, trailers
  server/         # listener, accept loop, handler wiring
```

## Requirements

- Go 1.25+

## Run the server

```bash
go run ./cmd/httpserver
```

It listens on port `42069`.

Stop it with Ctrl+C.

## Routes

| Path | What it does |
|------|----------------|
| `/` | HTML 200 |
| `/yourproblem` | HTML 400 |
| `/myproblem` | HTML 500 |
| `/video` | serves `assets/vim.mp4` if that file exists |
| `/httpbin/...` | proxies to `https://httpbingo.org/...` with chunked transfer + trailers |

## Try it

Normal request:

```bash
curl -i http://localhost:42069/
```

See raw chunked proxy output (curl hides the chunk framing):

```bash
echo -e "GET /httpbin/range/4096 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
```

Or stream:

```bash
echo -e "GET /httpbin/stream/100 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
```

The proxy announces `X-Content-SHA256` and `X-Content-Length` as trailers, then sends them after the body chunks.

## Tests

```bash
go test ./internal/...
```

## Notes

- Connections are closed after each response (no keep-alive yet).
- Chunk sizes are hex, like the HTTP/1.1 chunked encoding spec.
- `/video` needs `assets/vim.mp4` in the project root when you run the server.
