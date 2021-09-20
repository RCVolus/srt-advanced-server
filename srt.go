package main

import (
	"log"

	"github.com/haivision/srtgo"
)

func ListenIngressSocket() {
	options := make(map[string]string)
	options["transtype"] = "live"
	options["latency"] = "200"
	options["blocking"] = "0"

	sck := srtgo.NewSrtSocket("0.0.0.0", 11001, options)

	log.Println("Waiting for ingest SRT connections on 0.0.0.0:11001")

	// sck.SetListenCallback(ListenCallback)

	defer sck.Close()
	sck.Listen(10)

	for {
		HandleIngressSocket(sck)
	}
}

func HandleIngressSocket(sck *srtgo.SrtSocket) {
	socket, _, _ := sck.Accept()
	log.Println("New SRT socket for ingest opened")
	defer socket.Close()

	streamid, err := socket.GetSockOptString(srtgo.SRTO_STREAMID)
	if err != nil {
		log.Fatalln("Failed to get stream id, %s", err)
	}
	log.Println("Stream id: %s", streamid)

	for {
		// Create a new buffer
		// UDP packet cannot be larger than MTU (1500)
		buff := make([]byte, 1500)

		// 5s timeout
		n, err := socket.Read(buff)
		if err != nil {
			log.Println("Error occurred while reading SRT socket:", err)
			break
		}

		if n == 0 {
			// End of stream
			log.Println("Received no bytes, stopping stream.")
			break
		}

		// Send raw data to other streams
		buff = buff[:n]
		streamSender.Broadcast <- buff
	}

	// Close stream
	streamSender.Close()
}

/* func ListenCallback(socket *srtgo.SrtSocket, version int, addr *net.UDPAddr, streamid string) bool {
	fmt.Printf("Streamid: %v\n", streamid)

	return true
} */

func ListenEgressSocket() {
	// Open up a new socket
	options := make(map[string]string)
	options["transtype"] = "live"
	options["latency"] = "200"
	options["blocking"] = "0"

	sck := srtgo.NewSrtSocket("0.0.0.0", 11002, options)

	log.Println("Waiting for egest SRT connection on 0.0.0.0:11002")

	defer sck.Close()
	sck.Listen(10)

	for {
		socket, _, _ := sck.Accept()
		log.Println("New SRT socket for sending opened")

		// Send options
		/* sendSocket.SetSockOptInt(srtgo.SRTO_LATENCY, 2000)
		sendSocket.SetSockOptInt(srtgo.SRTO_RCVBUF, 4096)
		sendSocket.SetSockOptInt(srtgo.SRTO_SNDBUF, 4096) */

		// go SendSocket(*s)

		// Register new output
		go HandleEgressSocket(*socket)
	}
}

func HandleEgressSocket(socket srtgo.SrtSocket) {
	c := make(chan []byte, 1024)
	streamSender.Register(c)

	for data := range c {
		if len(data) < 1 {
			log.Println("SRT viewer removed, end of stream")
			break
		}

		_, err := socket.Write(data)
		if err != nil {
			log.Printf("Remove SRT viewer because of sending error, %s", err)
			break
		}
	}

	streamSender.Unregister(c)
	socket.Close()
}
