package srt

import (
	"log"
	"net"
	"time"

	"github.com/RCVolus/srt-advanced-server/stream"
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
		socket, addr, _ := sck.Accept()
		log.Println("New SRT socket for ingest opened")
		go HandleIngressSocket(socket, addr)
	}
}

func HandleIngressSocket(socket *srtgo.SrtSocket, addr *net.UDPAddr) {
	streamId, err := socket.GetSockOptString(srtgo.SRTO_STREAMID)
	if err != nil {
		log.Fatalln("Failed to get stream id, %s", err)
	}
	log.Println("Receiving new stream with StreamId ", streamId)

	ingestInfo := &stream.IngestStreamInformation{
		StreamId:    streamId,
		Remote:      addr.String(),
		ConnectedAt: time.Now(),
	}

	streamSender := stream.NewStreamSender(*ingestInfo)
	stream.IngestStreams[streamId] = *streamSender

	for {
		// Create a new buffer
		// UDP packet cannot be larger than MTU (1500)
		buff := make([]byte, 1500)

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
	socket.Close()

	delete(stream.IngestStreams, streamId)
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
		socket, addr, _ := sck.Accept()
		log.Println("New SRT socket for egest opened")

		// Send options
		/* sendSocket.SetSockOptInt(srtgo.SRTO_LATENCY, 2000)
		sendSocket.SetSockOptInt(srtgo.SRTO_RCVBUF, 4096)
		sendSocket.SetSockOptInt(srtgo.SRTO_SNDBUF, 4096) */

		// go SendSocket(*s)

		// Register new output
		go HandleEgressSocket(*socket, addr)
	}
}

func HandleEgressSocket(socket srtgo.SrtSocket, addr *net.UDPAddr) {
	c := make(chan []byte, 1024)
	stream, ok := stream.IngestStreams["test"]

	if !ok {
		log.Println("Unable to create SRT viewer, no ingest stream found")
		socket.SetRejectReason(srtgo.RejectionReasonNotFound)
		socket.Close()
		return
	}

	egestInfo := &stream.EgestStreamInformation{
		Remote:      addr.String(),
		ConnectedAt: time.Now(),
	}

	stream.Register(c)

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

	stream.Unregister(c)
	socket.Close()
}
