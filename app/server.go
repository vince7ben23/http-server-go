package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var dir = flag.String("directory", "/tmp/", "dir for file requests")

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

func (r *Request) isConnectClosed() bool {
	return strings.EqualFold(r.Headers["Connection"], "close")
}

type Server struct {
	Listener net.Listener
}

func (s *Server) Init() {
	s.initListener()

	for {
		conn := s.Accept()
		fmt.Printf("Connection establised with %v\n", conn.RemoteAddr())
		go handleRequest(conn)
	}
}

func (s *Server) initListener() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	s.Listener = l
}

func (s *Server) Accept() net.Conn {
	conn, err := s.Listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	return conn
}

type Response struct {
	Status  int
	Headers map[string]string
	Body    string
}

var statusText = map[int]string{200: "OK", 201: "Created", 404: "Not Found", 500: "Internal Server Error"}

func (r *Response) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "HTTP/1.1 %d %s\r\n", r.Status, statusText[r.Status])
	if r.Body != "" {
		r.Headers["Content-Length"] = strconv.Itoa(len(r.Body))
	}
	for k, v := range r.Headers {
		fmt.Fprintf(&sb, "%s: %s\r\n", k, v)
	}
	sb.WriteString("\r\n")
	sb.WriteString(r.Body)
	return sb.String()
}

func (r *Response) updateConnectionHeader(closeConn bool) {
	if closeConn {
		r.Headers["Connection"] = "close"
	} else {
		r.Headers["Connection"] = "keep-alive"
	}
}

type Route struct {
	Match   func(req *Request) bool
	Handler func(req *Request) *Response
}

var routes = []Route{
	{
		Match:   func(req *Request) bool { return req.Path == "/" },
		Handler: func(req *Request) *Response { return &Response{Status: 200, Headers: map[string]string{}} },
	},
	{
		Match: func(req *Request) bool { return strings.HasPrefix(req.Path, "/echo/") },
		Handler: func(req *Request) *Response {
			body := strings.TrimPrefix(req.Path, "/echo/")
			return &Response{Status: 200, Headers: map[string]string{"Content-Type": "text/plain"}, Body: body}
		},
	},
	{
		Match: func(req *Request) bool { return req.Path == "/user-agent" },
		Handler: func(req *Request) *Response {
			ua := req.Headers["User-Agent"]
			return &Response{Status: 200, Headers: map[string]string{"Content-Type": "text/plain"}, Body: ua}
		},
	},
	{
		Match: func(req *Request) bool { return req.Method == "GET" && strings.HasPrefix(req.Path, "/files/") },
		Handler: func(req *Request) *Response {
			filename := strings.TrimPrefix(req.Path, "/files/")
			filepath := *dir + filename
			content, err := os.ReadFile(filepath)
			if err != nil {
				return &Response{Status: 404, Headers: map[string]string{}}
			}
			return &Response{Status: 200, Headers: map[string]string{"Content-Type": "application/octet-stream"}, Body: string(content)}

		},
	},
	{
		Match: func(req *Request) bool { return req.Method == "POST" && strings.HasPrefix(req.Path, "/files/") },
		Handler: func(req *Request) *Response {
			filename := strings.TrimPrefix(req.Path, "/files/")
			filepath := *dir + filename
			err := os.WriteFile(filepath, req.Body, 0644)
			if err != nil {
				return &Response{Status: 500, Headers: map[string]string{}}
			}
			return &Response{Status: 201, Headers: map[string]string{}}
		},
	},
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		req, err := parseRequest(reader)
		if err != nil {
			if err == io.EOF {
				// Client closed the connection, server receives EOF
				return
			}
			fmt.Println("Error reading from request: ", err.Error())
			return
		}
		fmt.Printf("Request: \n%+v\n", req)

		response := generateResponseByRoute(req)
		closeConn := req.isConnectClosed()
		response.updateConnectionHeader(closeConn)

		fmt.Printf("Response: \n%s\n", response)
		conn.Write([]byte(response.String()))

		if closeConn {
			return
		}
	}

}

func parseRequest(reader *bufio.Reader) (*Request, error) {
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	parts := strings.Fields(requestLine)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid request line: %q", requestLine)
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Headers: make(map[string]string),
	}

	for {
		header, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if header == "\r\n" {
			break
		}
		if kv := strings.SplitN(header, ":", 2); len(kv) == 2 {
			req.Headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	// Read body if Content-Length is specified
	if cl, ok := req.Headers["Content-Length"]; ok {
		n, err := strconv.Atoi(cl)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %s", cl)
		}
		if n > 0 {
			req.Body = make([]byte, n)
			if _, err := io.ReadFull(reader, req.Body); err != nil {
				return nil, err
			}
		}
	}

	// fmt.Printf("Body: %s\n", string(req.Body))

	return req, nil
}

func generateResponseByRoute(req *Request) *Response {

	for _, route := range routes {
		if route.Match(req) {
			return route.Handler(req)
		}
	}
	return (&Response{Status: 404, Headers: map[string]string{}})
}

func main() {
	flag.Parse()
	server := &Server{}
	server.Init()

}
