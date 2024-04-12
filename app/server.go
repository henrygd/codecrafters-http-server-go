package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var directory = flag.String("directory", "/tmp", "a string")

func main() {
	flag.Parse()

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

	buf := make([]byte, 1024)
	reqLength, err := c.Read(buf)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		os.Exit(1)
	}
	reqString := string(buf[0:reqLength])
	reqPath := strings.Split(reqString, " ")[1]

	if reqPath == "/" {
		c.Write(createResponse("text/plain", "Hello"))
		return
	}

	if strings.HasPrefix(reqPath, "/echo/") {
		randStr := reqPath[6:]
		c.Write(createResponse("text/plain", randStr))
		return
	}

	if strings.HasPrefix(reqPath, "/user-agent") {
		splitReq := strings.Split(reqString, "User-Agent: ")
		agent := strings.Split(splitReq[1], "\r\n")[0]
		c.Write(createResponse("text/plain", agent))
		return
	}

	if strings.HasPrefix(reqPath, "/files/") {
		filename := reqPath[7:]
		dir := strings.TrimSuffix(*directory, "/") + "/"
		// GET /files
		if strings.HasPrefix(reqString, "GET") {
			fileContent, err := os.ReadFile(dir + filename)
			if err != nil {
				c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			c.Write(createResponse("application/octet-stream", string(fileContent)))
			return
		}
		// POST /files
		if strings.HasPrefix(reqString, "POST") {
			splitReq := strings.Split(reqString, "Content-Length: ")
			y := strings.Split(splitReq[1], "\r\n")[0]
			contentLength, err := strconv.Atoi(y)
			if err != nil {
				c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				return
			}
			bodyContent := reqString[reqLength-contentLength : reqLength]
			err = os.WriteFile(dir+filename, []byte(bodyContent), 0644)
			if err != nil {
				c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				return
			}
			c.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
			return
		}
	}

	c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func createResponse(contentType string, content string) []byte {
	res := []byte("HTTP/1.1 200 OK\r\n")
	res = append(res, []byte(fmt.Sprintf("Content-Type: %s\r\n", contentType))...)
	res = append(res, []byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content)))...)
	res = append(res, []byte(content)...)
	return res
}
