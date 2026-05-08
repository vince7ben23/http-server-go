package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buff := make([]byte, 1024)

	n, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}
	// fmt.Println("n:", n)
	request := string(buff[:n])
	fmt.Println("Received:", request)
	req_line := strings.Split(request, "\r\n")[0]
	path := strings.Split(req_line, " ")[1]

	var response string
	if path == "/" {
		response =
			"HTTP/1.1 200 OK\r\n" +
				"\r\n" +
				""
	} else {
		response =
			"HTTP/1.1 404 Not Found\r\n" +
				"\r\n" +
				""
	}

	conn.Write([]byte(response))
}
