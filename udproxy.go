package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error:", err)
	}
}

type udproxyConfig struct {
	Backends map[string]string `json:"backends"`
	Clients  map[string]string `json:"clients"`
}

func backend(local, remote string, quit chan struct{}, input chan []byte) {
	laddr, err := net.ResolveUDPAddr("udp", local)
	checkErr(err)

	raddr, err := net.ResolveUDPAddr("udp", remote)
	checkErr(err)

	conn, err := net.DialUDP("udp", laddr, raddr)
	checkErr(err)

	defer conn.Close()

	for {
		select {
		case <-quit:
			return
		case msg := <-input:
			_, err := conn.Write(msg)
			checkErr(err)
		}
	}
}

func spawnBackend(local, remote string) (chan struct{}, chan []byte) {
	quit := make(chan struct{})
	input := make(chan []byte)

	go backend(local, remote, quit, input)

	return quit, input
}

func main() {
	var config udproxyConfig

	if len(os.Args) < 2 {
		log.Fatalln("Usage:", os.Args[0], "<config file>")
	}

	data, err := ioutil.ReadFile(os.Args[1])
	checkErr(err)

	err = yaml.Unmarshal(data, &config)
	checkErr(err)
}
