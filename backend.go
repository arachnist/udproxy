package main

import (
	"log"
	"net"
)

func backend(local, remote string, quit chan struct{}, input chan []byte) {
	laddr, err := net.ResolveUDPAddr("udp", local)
	checkErr(err)

	raddr, err := net.ResolveUDPAddr("udp", remote)
	checkErr(err)

	conn, err := net.DialUDP("udp", laddr, raddr)
	checkErr(err)

	defer conn.Close()

	log.Println("Backend connection ready for action", local, remote)

	for {
		select {
		case <-quit:
			return
		case msg := <-input:
			_, err := conn.Write(msg)
			if err != nil {
				log.Println("Couldn't send message:", err)
			}
		}
	}
}

func spawnBackend(local, remote string) (chan struct{}, chan []byte) {
	quit := make(chan struct{})
	input := make(chan []byte)

	go backend(local, remote, quit, input)

	return quit, input
}
