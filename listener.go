package main

import (
	"log"
	"net"
	"time"
)

func listener(listen string, quit chan struct{}, dispatcher func(net.IP, []byte)) {
	buf := make([]byte, 50000)

	laddr, err := net.ResolveUDPAddr("udp", listen)
	checkErr(err)

	conn, err := net.ListenUDP("udp", laddr)
	checkErr(err)

	defer conn.Close()
	log.Println("Listener ready for action", listen)

	for {
		select {
		case <-quit:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(90 * time.Millisecond))
			n, addr, err := conn.ReadFromUDP(buf)
			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}
			if err != nil {
				log.Println("Error:", err)
			}
			go dispatcher(addr.IP, buf[0:n])
		}
	}
}

func spawnListener(listen string, dispatcher func(net.IP, []byte)) chan struct{} {
	quit := make(chan struct{})

	go listener(listen, quit, dispatcher)

	return quit
}
