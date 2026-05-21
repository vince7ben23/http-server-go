# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a CodeCrafters challenge: building an HTTP/1.1 server from scratch in Go. All implementation lives in `app/server.go`. Submit by pushing to `master` — CodeCrafters runs tests automatically.

## Commands

```sh
# Run the server (builds and executes)
./your_server.sh

# Run with a file-serving directory
./your_server.sh --directory /tmp/

# Build only
go build ./app/...

# Run unit tests
go test ./app/...

# Test by curling the running server
curl -v http://localhost:4221/
```

The server listens on port `4221`.

## Architecture

Single file: `app/server.go`. Key types:

- **`Server`** — wraps `net.Listener`; `Init()` accepts TCP connections and spawns `go handleRequest(conn)` per connection with a keep-alive loop.
- **`Request`** — parsed HTTP request: `Method`, `Path`, `Version`, `Headers` (map), `Body` ([]byte read via `Content-Length`).
  - `isConnectClosed()` — returns true if `Connection: close` header is set
  - `isAcceptEncoding()` — returns true if `Accept-Encoding` contains `gzip`
- **`Response`** — `Status`, `Headers` (map), `Body` ([]byte). `String()` serializes to HTTP/1.1 wire format.
  - `updateConnectionHeader()` — sets `Connection: close`
  - `updateContentEncoding()` — sets `Content-Encoding: gzip`
  - `gzipBody()` — compresses `Body` in place with gzip
- **`Route`** — a `Match` predicate and a `Handler` func, both taking `*Request`. Routes are matched in order from the `routes` slice; first match wins, 404 if none match.

Status codes in `statusText`: 200, 201, 404, 500.

## Routes

| Method | Path | Behaviour |
|--------|------|-----------|
| GET | `/` | 200 empty body |
| GET | `/echo/{text}` | 200, body = text after `/echo/` |
| GET | `/user-agent` | 200, body = User-Agent header value |
| GET | `/files/{name}` | 200 + file content, or 404 if not found |
| POST | `/files/{name}` | Write request body to file, 201 or 500 |

The `--directory` flag sets the base path for `/files/` routes (GET reads, POST writes).

## Compression

If a request includes `Accept-Encoding: gzip`, the response body is compressed with gzip and `Content-Encoding: gzip` is set. Logic lives in `handleRequest()` via `isAcceptEncoding()`, `gzipBody()`, and `updateContentEncoding()`.

## Adding New Routes

Append a `Route` to the `routes` slice in `server.go`. The pattern is:

```go
{
    Match:   func(req *Request) bool { return /* condition */ },
    Handler: func(req *Request) *Response { return &Response{...} },
},
```

To support a new HTTP status code, add it to the `statusText` map.

## Testing

29 test cases in `app/server_test.go` covering: request parsing, response serialisation, all 5 routes, `Request`/`Response` helper methods, gzip compression, and file I/O.
