package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRequest(t *testing.T) {
	t.Run("basic GET", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
		req, err := parseRequest(bufio.NewReader(strings.NewReader(raw)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Method != "GET" {
			t.Errorf("Method = %q, want %q", req.Method, "GET")
		}
		if req.Path != "/" {
			t.Errorf("Path = %q, want %q", req.Path, "/")
		}
		if req.Version != "HTTP/1.1" {
			t.Errorf("Version = %q, want %q", req.Version, "HTTP/1.1")
		}
		if req.Headers["Host"] != "localhost" {
			t.Errorf("Host header = %q, want %q", req.Headers["Host"], "localhost")
		}
	})

	t.Run("GET with User-Agent header", func(t *testing.T) {
		raw := "GET /user-agent HTTP/1.1\r\nUser-Agent: test-agent\r\n\r\n"
		req, err := parseRequest(bufio.NewReader(strings.NewReader(raw)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Headers["User-Agent"] != "test-agent" {
			t.Errorf("User-Agent = %q, want %q", req.Headers["User-Agent"], "test-agent")
		}
	})

	t.Run("POST with body", func(t *testing.T) {
		body := "hello"
		raw := "POST /files/test.txt HTTP/1.1\r\nContent-Length: 5\r\n\r\n" + body
		req, err := parseRequest(bufio.NewReader(strings.NewReader(raw)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Method != "POST" {
			t.Errorf("Method = %q, want %q", req.Method, "POST")
		}
		if string(req.Body) != body {
			t.Errorf("Body = %q, want %q", string(req.Body), body)
		}
	})
}

func TestResponseString(t *testing.T) {
	t.Run("200 with body", func(t *testing.T) {
		resp := &Response{Status: 200, Headers: map[string]string{"Content-Type": "text/plain"}, Body: "hello"}
		s := resp.String()
		if !strings.HasPrefix(s, "HTTP/1.1 200 OK\r\n") {
			t.Errorf("response does not start with status line, got: %q", s[:30])
		}
		if !strings.Contains(s, "Content-Length: 5") {
			t.Errorf("response missing Content-Length: 5, got: %q", s)
		}
		if !strings.HasSuffix(s, "\r\nhello") {
			t.Errorf("response body incorrect, got: %q", s)
		}
	})

	t.Run("404 no body", func(t *testing.T) {
		resp := &Response{Status: 404, Headers: map[string]string{}}
		s := resp.String()
		if !strings.HasPrefix(s, "HTTP/1.1 404 Not Found\r\n") {
			t.Errorf("response does not start with 404 status line, got: %q", s[:30])
		}
		if !strings.Contains(s, "Content-Length: 0") {
			t.Errorf("response missing Content-Length: 0, got: %q", s)
		}
	})
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		name             string
		req              *Request
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:       "GET / → 200",
			req:        &Request{Method: "GET", Path: "/", Headers: map[string]string{}},
			wantStatus: 200,
		},
		{
			name:             "GET /echo/hello → 200 body=hello",
			req:              &Request{Method: "GET", Path: "/echo/hello", Headers: map[string]string{}},
			wantStatus:       200,
			wantBodyContains: "hello",
		},
		{
			name:             "GET /user-agent → 200 body=agent string",
			req:              &Request{Method: "GET", Path: "/user-agent", Headers: map[string]string{"User-Agent": "test-agent"}},
			wantStatus:       200,
			wantBodyContains: "test-agent",
		},
		{
			name:       "GET /nonexistent → 404",
			req:        &Request{Method: "GET", Path: "/nonexistent", Headers: map[string]string{}},
			wantStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := generateResponseByRoute(tt.req)
			if resp.Status != tt.wantStatus {
				t.Errorf("Status = %d, want %d", resp.Status, tt.wantStatus)
			}
			if tt.wantBodyContains != "" && !strings.Contains(resp.Body, tt.wantBodyContains) {
				t.Errorf("Body = %q, want it to contain %q", resp.Body, tt.wantBodyContains)
			}
		})
	}
}

func TestRequestHelpers(t *testing.T) {
	t.Run("isConnectClosed true", func(t *testing.T) {
		req := &Request{Headers: map[string]string{"Connection": "close"}}
		if !req.isConnectClosed() {
			t.Error("expected isConnectClosed() = true")
		}
	})

	t.Run("isConnectClosed false when keep-alive", func(t *testing.T) {
		req := &Request{Headers: map[string]string{"Connection": "keep-alive"}}
		if req.isConnectClosed() {
			t.Error("expected isConnectClosed() = false")
		}
	})

	t.Run("isConnectClosed false when absent", func(t *testing.T) {
		req := &Request{Headers: map[string]string{}}
		if req.isConnectClosed() {
			t.Error("expected isConnectClosed() = false when header absent")
		}
	})

	t.Run("isConnectClosed case-insensitive", func(t *testing.T) {
		req := &Request{Headers: map[string]string{"Connection": "Close"}}
		if !req.isConnectClosed() {
			t.Error("expected isConnectClosed() = true for 'Close'")
		}
	})

	t.Run("isAcceptEncoding true for gzip", func(t *testing.T) {
		req := &Request{Headers: map[string]string{"Accept-Encoding": "gzip"}}
		if !req.isAcceptEncoding() {
			t.Error("expected isAcceptEncoding() = true")
		}
	})

	t.Run("isAcceptEncoding true for multiple schemes with gzip", func(t *testing.T) {
		req := &Request{Headers: map[string]string{"Accept-Encoding": "deflate, gzip, br"}}
		if !req.isAcceptEncoding() {
			t.Error("expected isAcceptEncoding() = true for 'deflate, gzip, br'")
		}
	})

	t.Run("isAcceptEncoding false when absent", func(t *testing.T) {
		req := &Request{Headers: map[string]string{}}
		if req.isAcceptEncoding() {
			t.Error("expected isAcceptEncoding() = false when header absent")
		}
	})

	t.Run("isAcceptEncoding false for deflate only", func(t *testing.T) {
		req := &Request{Headers: map[string]string{"Accept-Encoding": "deflate"}}
		if req.isAcceptEncoding() {
			t.Error("expected isAcceptEncoding() = false for 'deflate'")
		}
	})
}

func TestResponseHelpers(t *testing.T) {
	t.Run("updateConnectionHeader sets close", func(t *testing.T) {
		resp := &Response{Status: 200, Headers: map[string]string{}}
		resp.updateConnectionHeader()
		if resp.Headers["Connection"] != "close" {
			t.Errorf("Connection = %q, want %q", resp.Headers["Connection"], "close")
		}
	})

	t.Run("updateContentEncoding sets gzip", func(t *testing.T) {
		resp := &Response{Status: 200, Headers: map[string]string{}}
		resp.updateContentEncoding()
		if resp.Headers["Content-Encoding"] != "gzip" {
			t.Errorf("Content-Encoding = %q, want %q", resp.Headers["Content-Encoding"], "gzip")
		}
	})

	t.Run("gzipBody compresses body", func(t *testing.T) {
		resp := &Response{Status: 200, Headers: map[string]string{}, Body: "hello"}
		if err := resp.gzipBody(); err != nil {
			t.Fatalf("gzipBody() error: %v", err)
		}

		gr, err := gzip.NewReader(bytes.NewReader([]byte(resp.Body)))
		if err != nil {
			t.Fatalf("gzip.NewReader error: %v", err)
		}
		decompressed, err := io.ReadAll(gr)
		if err != nil {
			t.Fatalf("ReadAll error: %v", err)
		}
		if string(decompressed) != "hello" {
			t.Errorf("decompressed body = %q, want %q", string(decompressed), "hello")
		}
	})

	t.Run("gzipBody on empty body", func(t *testing.T) {
		resp := &Response{Status: 200, Headers: map[string]string{}, Body: ""}
		if err := resp.gzipBody(); err != nil {
			t.Fatalf("gzipBody() error: %v", err)
		}

		gr, err := gzip.NewReader(bytes.NewReader([]byte(resp.Body)))
		if err != nil {
			t.Fatalf("gzip.NewReader error: %v", err)
		}
		decompressed, err := io.ReadAll(gr)
		if err != nil {
			t.Fatalf("ReadAll error: %v", err)
		}
		if string(decompressed) != "" {
			t.Errorf("decompressed body = %q, want empty", string(decompressed))
		}
	})
}

func TestFileRoutes(t *testing.T) {
	tmpDir := t.TempDir() + "/"
	*dir = tmpDir

	t.Run("POST /files/test.txt → 201", func(t *testing.T) {
		req := &Request{Method: "POST", Path: "/files/test.txt", Headers: map[string]string{}, Body: []byte("hello")}
		resp := generateResponseByRoute(req)
		if resp.Status != 201 {
			t.Errorf("Status = %d, want 201", resp.Status)
		}
		content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
		if err != nil {
			t.Fatalf("file was not created: %v", err)
		}
		if string(content) != "hello" {
			t.Errorf("file content = %q, want %q", string(content), "hello")
		}
	})

	t.Run("GET /files/test.txt → 200", func(t *testing.T) {
		req := &Request{Method: "GET", Path: "/files/test.txt", Headers: map[string]string{}}
		resp := generateResponseByRoute(req)
		if resp.Status != 200 {
			t.Errorf("Status = %d, want 200", resp.Status)
		}
		if resp.Body != "hello" {
			t.Errorf("Body = %q, want %q", resp.Body, "hello")
		}
	})

	t.Run("GET /files/missing.txt → 404", func(t *testing.T) {
		req := &Request{Method: "GET", Path: "/files/missing.txt", Headers: map[string]string{}}
		resp := generateResponseByRoute(req)
		if resp.Status != 404 {
			t.Errorf("Status = %d, want 404", resp.Status)
		}
	})
}
