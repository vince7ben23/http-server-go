package main

import (
	"fmt"
	// Uncomment this block to pass the first stage
	"net"
	"os"
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
	fmt.Println("Received:", string(buff[0:n]))

	response := "HTTP/1.1 200 OK\r\n" +
		// "Content-Length: 6\r\n" +
		// "Content-Type: text/plain\r\n" +
		"\r\n" +
		""
	conn.Write([]byte(response))
}
