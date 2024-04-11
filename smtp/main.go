package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <your-gmail> <receiver-email>", os.Args[0])
		os.Exit(1)
	}

	gmailServer := "smtp.gmail.com:587"

	tcpAddr, err := net.ResolveTCPAddr("tcp", gmailServer)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	var resp []byte
	var n int
	resp, n = recvFromServer(conn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "220")

	sendToServer(conn, "EHLO Alice\r\n")
	resp, n = recvFromServer(conn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "250")

	sendToServer(conn, "STARTTLS\r\n")
	resp, n = recvFromServer(conn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "220")

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: "smtp.gmail.com",
	})

	sendToServer(tlsConn, "EHLO Alice\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "250")

	sendToServer(tlsConn, "AUTH LOGIN\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "334")

	username := base64.StdEncoding.EncodeToString([]byte(os.Args[1]))
	sendToServer(tlsConn, username+"\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "334")

	fmt.Print("Enter Password: ")
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	checkError(err)

	password = base64.StdEncoding.EncodeToString([]byte(password))
	sendToServer(tlsConn, password+"\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "235")

	fromEmail := os.Args[1]
	sendToServer(tlsConn, "MAIL FROM: <"+fromEmail+">\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "250")

	toEmail := os.Args[2]
	sendToServer(tlsConn, "RCPT TO: <"+toEmail+">\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "250")

	sendToServer(tlsConn, "DATA\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "354")

	text := "Я люблю компьютерные сети!"
	message := "From: " + fromEmail + "\r\n"
	message += "To: " + toEmail + "\r\n"
	message += "Subject: " + text + "\r\n\r\n"
	message += text + "\r\n"
	sendToServer(tlsConn, message)

	sendToServer(tlsConn, ".\r\n")
	resp, n = recvFromServer(tlsConn)
	fmt.Println(string(resp[0:n]))
	checkResponse(resp[0:], "250")

	sendToServer(tlsConn, "QUIT\r\n")

	tlsConn.Close()

	os.Exit(0)
}

func sendToServer(conn net.Conn, data string) {
	_, err := conn.Write([]byte(data))
	checkError(err)
}

func recvFromServer(conn net.Conn) ([]byte, int) {
	var resp [512]byte

	n, err := conn.Read(resp[0:])
	checkError(err)

	return resp[0:], n
}

func checkResponse(resp []byte, code string) {
	if string(resp[0:3]) != code {
		fmt.Fprintf(os.Stderr, "%s reply not received from server.", code)
		os.Exit(1)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
