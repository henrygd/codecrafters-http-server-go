package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	fmt.Println("Listening on 0.0.0.0:4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

func handleRequest(c net.Conn) {
	defer c.Close()

	req := make([]byte, 1024)
	_, err := c.Read(req)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		os.Exit(1)
	}
	reqPath := strings.Split(string(req), " ")[1]

	var res []byte

	if !strings.HasPrefix(reqPath, "/echo/") {
		res = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
		c.Write(res)
		return
	}

	randStr := reqPath[6:]
	res = []byte("HTTP/1.1 200 OK\r\n")
	res = append(res, []byte("Content-Type: text/plain\r\n")...)
	res = append(res, []byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(randStr)))...)
	res = append(res, []byte(randStr)...)

	c.Write(res)
}
