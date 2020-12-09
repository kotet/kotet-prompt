package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
)

var socket_file string = "/tmp/kotet-prompt.sock"

func main() {
	listener, err := net.Listen("unix", socket_file)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer end()
	defer_chan := make(chan os.Signal, 1)
	signal.Notify(defer_chan, os.Interrupt)
	go func() {
		for signal := range defer_chan {
			close(defer_chan)
			log.Println(signal.String())
			end()
			os.Exit(130)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		go func() {
			err := exec.Command("play", "-q", "-n", "synth", "sin", "1000", "trim", "0", "0.05", "vol", "0.1").Run()
			if err != nil {
				log.Println(err.Error())
			}
		}()
		conn.Close()
	}
}

func end() {
	err := os.Remove(socket_file)
	if err != nil {
		log.Fatal(err.Error())
	}
}
