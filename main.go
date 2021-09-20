package main

import (
	"log"
)

var streamSender = newStreamSender()

func main() {
	go startHttp()

	go ListenEgressSocket()
	log.Println("Waiting for connections")
	ListenIngressSocket()

	// s.GetSockOptString()
	//....
}
