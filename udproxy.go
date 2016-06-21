package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
)

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error:", err)
	}
}

type Backend struct {
	Upstream string `json:"backend"`
	Local    string `json:"local"`
	input    chan []byte
	quit     chan struct{}
}

type Listener struct {
	Address string `json:"address"`
	quit    chan struct{}
}

type udproxyConfig struct {
	Backends map[string]Backend `json:"backends,inline"`
	Listen   []Listener         `json:"listen"`
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

	log.Println("Backend connection ready for action", local, remote)

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
	quit := make(chan struct{}, 1)

	if len(os.Args) < 2 {
		log.Fatalln("Usage:", os.Args[0], "<config file>")
	}

	data, err := ioutil.ReadFile(os.Args[1])
	checkErr(err)

	err = yaml.Unmarshal(data, &config)
	checkErr(err)

	data, err = yaml.Marshal(config)
	checkErr(err)
	log.Print("Parsed configuration:\n", string(data))

	for name, backend := range config.Backends {
		quit, input := spawnBackend(backend.Local, backend.Upstream)
		config.Backends[name] = Backend{
			Upstream: backend.Upstream,
			Local:    backend.Local,
			input:    input,
			quit:     quit,
		}
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		<-c
		log.Println("Shutting down!")
		for name, backend := range config.Backends {
			log.Println("Stopping backend", name)
			backend.quit <- struct{}{}
		}

		log.Println("Writing config!")
		data, err = yaml.Marshal(config)
		checkErr(err)

		err = ioutil.WriteFile(os.Args[1], data, 0600)
		checkErr(err)

		log.Println("Finally stopping!")
		quit <- struct{}{}
	}()

	<-quit
}
