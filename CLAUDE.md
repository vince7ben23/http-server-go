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

# Test by curling the running server
curl -v http://localhost:4221/
```

The server listens on port `4221`.

## Architecture

Single file: `app/server.go`. Key types:

- **`Server`** — wraps `net.Listener`, accepts TCP connections on `0.0.0.0:4221`, spawns a goroutine per connection via `go handleRequest(conn)`.
- **`Request`** — parsed HTTP request: `Method`, `Path`, `Headers` (map), `Body` ([]byte read via `Content-Length`).
- **`Response`** — status code, headers map, string body. `String()` serializes to HTTP/1.1 wire format.
- **`Route`** — a `Match` predicate and a `Handler` func, both taking `*Request`. Routes are matched in order from the `routes` slice; first match wins, 404 if none match.

The `--directory` flag sets the base path for `/files/` routes (GET reads, POST writes).

## Adding New Routes

Append a `Route` to the `routes` slice in `server.go`. The pattern is:

```go
{
    Match:   func(req *Request) bool { return /* condition */ },
    Handler: func(req *Request) *Response { return &Response{...} },
},
```

To support a new HTTP status code, add it to the `statusText` map.
