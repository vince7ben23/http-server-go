package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type Server struct {
	listener net.Listener
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
	s.listener = l
}

func (s *Server) Accept() net.Conn {
	conn, err := s.listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	return conn
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var headers []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}
		headers = append(headers, line)
		if line == "\r\n" {
			break
		}
	}

	req_line := headers[0]
	fmt.Printf("Request:\n%s\n", req_line)
	parts := strings.Split(req_line, " ")
	path := parts[1]

	var response string
	switch {

	case path == "/":
		response =
			"HTTP/1.1 200 OK\r\n\r\n"

	case strings.HasPrefix(path, "/echo/"):
		path_parts := strings.Split(path, "/")
		echo := path_parts[2]
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s", len(echo), echo)

	case path == "/user-agent":
		var user_agent string
		for i, v := range headers {
			if i == 0 {
				continue
			}
			if strings.Contains(v, "User-Agent:") {
				user_agent = strings.TrimSpace(strings.Split(v, ":")[1])
			}
		}
		response = fmt.Sprintf(
			"HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/plain\r\n"+
				"Content-Length: %d\r\n"+
				"\r\n"+
				"%s", len(user_agent), user_agent)

	default:
		response =
			"HTTP/1.1 404 Not Found\r\n\r\n"
	}

	fmt.Printf("Response:\n%s\n", response)

	conn.Write([]byte(response))
}

func main() {
	server := &Server{}
	server.Init()

}
