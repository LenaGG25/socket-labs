package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	proxyAddr = flag.String("addr", ":8000", "proxy server address")
)

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", *proxyAddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	log.Println("proxy server listen on", *proxyAddr)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("handleConn panic:", err)
		}
	}()
	log.Printf("Received a connection from %s\n", conn.RemoteAddr())
	defer conn.Close()
	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		log.Println("read error: ", err)
		return
	}
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		log.Println("invalid request line:", line)
		return
	}
	method, url, version := parts[0], parts[1], parts[2]
	log.Printf("method: %s, url: %s, version: %s", method, url, version)

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("read error:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	filename := strings.SplitN(url, "/", 2)[1]
	log.Println("filename:", filename)

	if _, err := os.Stat(filename); err == nil || os.IsExist(err) {
		log.Println("File found in cache")

		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		conn.Write([]byte("Content-Type: text/html\r\n"))

		conn.Write([]byte("\r\n"))

		f, err := os.OpenFile(filename, os.O_RDONLY, 0666)
		if err != nil {
			log.Println("open file error:", err)
			return
		}
		defer f.Close()
		bs, err := io.ReadAll(f)
		if err != nil {
			log.Println("read file error:", err)
			return
		}

		conn.Write(bs)
		conn.Write([]byte("\r\n"))
		log.Println("Read from cache")
	} else {
		log.Println("File not found in cache")

		hostn := strings.Replace(filename, "www.", "", 1)
		log.Println("hostn:", hostn)

		c, err := net.Dial("tcp", hostn+":80")
		if err != nil {
			log.Println("dial error:", err)
			return
		}
		defer c.Close()

		c.Write([]byte("GET " + "http://" + filename + " HTTP/1.1\r\n"))
		c.Write([]byte("Host: " + hostn + ":80" + "\r\n"))
		c.Write([]byte("Connection: close\r\n"))
		c.Write([]byte("\r\n"))
		log.Println("Write to server")

		bs, err := io.ReadAll(c)
		if err != nil {
			log.Println("read error:", err)
			return
		}

		tmpFile, err := os.OpenFile("./"+filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println("open error:", err)
			return
		}
		defer tmpFile.Close()

		s := string(bs)
		body := strings.Split(s, "\r\n\r\n")[1]
		tmpFile.Write([]byte(body))

		c.Write(bs)
		log.Println("Read from server")
	}
}
