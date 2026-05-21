# HTTP Server (Go)

A from-scratch HTTP/1.1 server implemented in Go.

## Features

- TCP server on `0.0.0.0:4221`, one goroutine per connection
- HTTP/1.1 keep-alive (persistent connections)
- gzip response compression (via `Accept-Encoding: gzip`)
- Static file serving (`GET /files/` and `POST /files/`)
- `GET /echo/{text}` — echoes path segment back
- `GET /user-agent` — returns the `User-Agent` header value

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | 200 OK, empty body |
| GET | `/echo/{text}` | 200, body = `{text}` |
| GET | `/user-agent` | 200, body = User-Agent value |
| GET | `/files/{name}` | 200 + file content, or 404 |
| POST | `/files/{name}` | Write body to file, 201 or 500 |

## Running

```sh
# Build and start the server
./your_server.sh

# With file-serving directory
./your_server.sh --directory /tmp/

# Build only
go build ./app/...

# Run tests
go test ./app/...
```
