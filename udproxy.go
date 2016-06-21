package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error:", err)
	}
}

type Backend struct {
	Upstream string        `json:"backend"`
	Local    string        `json:"local"`
	input    chan []byte   `json:"input"`
	quit     chan struct{} `json:"quit"`
}

type udproxyConfig struct {
	Backends map[string]Backend `json:"backends,inline"`
	Sockets  []string           `json:"sockets"`
	Clients  map[string]string  `json:"clients"`
}

func backend(local, remote string, quit chan struct{}, input chan []byte) {
	laddr, err := net.ResolveUDPAddr("udp", local)
	checkErr(err)

	raddr, err := net.ResolveUDPAddr("udp", remote)
	checkErr(err)

	conn, err := net.DialUDP("udp", laddr, raddr)
	checkErr(err)

	defer conn.Close()

	log.Println("Ready for action", local, remote)

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

	for name, backend := range config.Backends {
		quit, input := spawnBackend(backend.Local, backend.Upstream)
		config.Backends[name] = Backend{
			Upstream: backend.Upstream,
			Local:    backend.Local,
			input:    input,
			quit:     quit,
		}
	}

	data, err = yaml.Marshal(config)
	fmt.Println(string(data))

	time.Sleep(90 * time.Second)
}
