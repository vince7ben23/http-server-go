package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
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

	return req, nil
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	req, err := parseRequest(reader)
	fmt.Printf("Request: \n%+v\n", req)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}

	response := generateResponse(req)
	fmt.Printf("Response: \n%s\n", response)
	conn.Write([]byte(response))
}

func generateResponse(req *Request) string {
	switch {
	case req.Path == "/":
		return "HTTP/1.1 200 OK\r\n\r\n"
	case strings.HasPrefix(req.Path, "/echo/"):
		echo := strings.TrimPrefix(req.Path, "/echo/")
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	case req.Path == "/user-agent":
		ua := req.Headers["User-Agent"]
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(ua), ua)
	default:
		return "HTTP/1.1 404 Not Found\r\n\r\n"
	}
}

func main() {
	server := &Server{}
	server.Init()

}
