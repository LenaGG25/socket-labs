package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

func main() {
	l, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 12000,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %s\n", l.LocalAddr().String())
	defer l.Close()
	var buf [1024]byte
	for {
		n, remoteAddr, err := l.ReadFromUDP(buf[0:])
		if err != nil {
			log.Println(err)
		}
		random := rand.Intn(10)
		if random < 4 {
			continue
		}
		data := string(buf[:n])
		data = strings.TrimRight(data, "\n")
		data = strings.ToUpper(data)

		log.Printf("%s: %s\n", remoteAddr, data)
		time.Sleep(time.Duration(random) * time.Millisecond)
		l.WriteToUDP([]byte(fmt.Sprintf("%s\n", data)), remoteAddr)
	}

}
