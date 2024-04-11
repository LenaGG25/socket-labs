package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	service := ":6789"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	fmt.Println("Готов к обслуживанию на порту 6789 ...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		handleRequest(conn)
		conn.Close()
	}
}

func handleRequest(conn net.Conn) {
	var reqMsg [512]byte

	n, err := conn.Read(reqMsg[0:])
	if err != nil {
		return
	}

	reqLine := bytes.Split(reqMsg[0:n], []byte("\n"))[0]
	fmt.Println(string(reqLine))

	fileURL := strings.Fields(string(reqLine))[1]
	fileName := fileURL[1:]

	file, err := os.Open(fileName)
	defer file.Close()

	if err != nil {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	} else {
		_, err2 := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err2 != nil {
			return
		}

		chunk := make([]byte, 4)
		for {
			n, err2 = file.Read(chunk)
			if err2 == io.EOF {
				break
			}

			_, err2 = conn.Write(chunk[:n])
			if err2 != nil {
				return
			}
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
