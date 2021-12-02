package srt

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/RCVolus/srt-advanced-server/stream"
	"github.com/haivision/srtgo"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	SrtStats = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "srt_stats",
		Help: "SRT stats provided by the SRT library.",
	}, []string{"stream", "stat", "type", "remoteAddr"})
)

func ListenIngressSocket(port uint16) {
	options := make(map[string]string)
	options["transtype"] = "live"
	options["latency"] = "200"
	options["blocking"] = "1"

	sck := srtgo.NewSrtSocket("0.0.0.0", port, options)

	log.Println(fmt.Sprintf("Waiting for ingest SRT connections on 0.0.0.0:%d", port))

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
		log.Fatalln(fmt.Sprintf("Failed to get stream id, %s", err))
	}
	log.Println(fmt.Sprintf("Receiving new stream with StreamId %s", streamId))

	ingestInfo := &stream.IngestStreamInformation{
		StreamId:    streamId,
		Remote:      addr.String(),
		ConnectedAt: time.Now(),
	}

	streamSender := stream.NewStreamSender(*ingestInfo, *socket)
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

		// Update stats
		UpdateStats(socket, streamId, "ingest", addr.String())
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

func ListenEgressSocket(port uint16, streamId string) {
	// Open up a new socket
	options := make(map[string]string)
	options["transtype"] = "live"
	options["latency"] = "200"
	options["blocking"] = "0"

	sck := srtgo.NewSrtSocket("0.0.0.0", port, options)

	log.Println(fmt.Sprintf("Waiting for egest SRT connection on 0.0.0.0:%d", port))

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
		go HandleEgressSocket(socket, addr, streamId)
	}
}

func HandleEgressSocket(socket *srtgo.SrtSocket, addr *net.UDPAddr, streamId string) {
	c := make(chan []byte, 1024)
	currentStream, ok := stream.IngestStreams[streamId]

	if !ok {
		log.Println("Unable to create SRT viewer, no ingest stream found")
		socket.SetRejectReason(srtgo.RejectionReasonNotFound)
		socket.Close()
		return
	}

	egestInfo := stream.EgestStreamInformation{
		Remote:      addr.String(),
		ConnectedAt: time.Now(),
	}

	currentStream.Register(c, egestInfo, socket)

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

		// Update stats
		UpdateStats(socket, "test", "egest", addr.String())
	}

	currentStream.Unregister(c)
	socket.Close()
}

func UpdateStats(socket *srtgo.SrtSocket, streamid string, streamType string, remoteAddr string) {
	// stats, _ := socket.Stats()

	/* v := reflect.ValueOf(stats)

	for i := 0; i < v.NumField(); i++ {
		// values[i] = v.Field(i).Interface()
		var metricValue float64

		switch v.Field(i).Type().Key().Kind() {
		case reflect.Int:
			metricValue = float64(v.Field(i).Interface().(int))
		case reflect.Int64:
			metricValue = float64(v.Field(i).Interface().(int64))
		case reflect.Float64:
			metricValue = v.Field(i).Interface().(float64)
		}

		SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": v.Field(i).Type().Name()}).Set(metricValue)
	} */

	/* SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MbpsBandwidth"}).Set(stats.MbpsBandwidth)
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MbpsMaxBW"}).Set(stats.MbpsMaxBW)
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MbpsRecvRate"}).Set(stats.MbpsRecvRate)
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MbpsSendRate"}).Set(stats.MbpsSendRate)
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MsRTT"}).Set(stats.MsRTT)
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvAvgBelatedTime"}).Set(stats.PktRcvAvgBelatedTime)
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "UsPktSndPeriod"}).Set(stats.UsPktSndPeriod)
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteAvailRcvBuf"}).Set(float64(stats.ByteAvailRcvBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteAvailSndBuf"}).Set(float64(stats.ByteAvailSndBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteMSS"}).Set(float64(stats.ByteMSS))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRcvBuf"}).Set(float64(stats.ByteRcvBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRcvDrop"}).Set(float64(stats.ByteRcvDrop))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRcvDropTotal"}).Set(float64(stats.ByteRcvDropTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRcvLoss"}).Set(float64(stats.ByteRcvLoss))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRcvLossTotal"}).Set(float64(stats.ByteRcvLossTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRcvUndecrypt"}).Set(float64(stats.ByteRcvUndecrypt))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRcvUndecryptTotal"}).Set(float64(stats.ByteRcvUndecryptTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRecv"}).Set(float64(stats.ByteRecv))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRecvTotal"}).Set(float64(stats.ByteRecvTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRetrans"}).Set(float64(stats.ByteRetrans))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteRetransTotal"}).Set(float64(stats.ByteRetransTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteSent"}).Set(float64(stats.ByteSent))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteSentTotal"}).Set(float64(stats.ByteSentTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteSndBuf"}).Set(float64(stats.ByteSndBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteSndDrop"}).Set(float64(stats.ByteSndDrop))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "ByteSndDropTotal"}).Set(float64(stats.ByteSndDropTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MsRcvBuf"}).Set(float64(stats.MsRcvBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MsRcvTsbPdDelay"}).Set(float64(stats.MsRcvTsbPdDelay))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MsSndBuf"}).Set(float64(stats.MsSndBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MsSndTsbPdDelay"}).Set(float64(stats.MsSndTsbPdDelay))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "MsTimeStamp"}).Set(float64(stats.MsTimeStamp))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktCongestionWindow"}).Set(float64(stats.PktCongestionWindow))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktFlightSize"}).Set(float64(stats.PktFlightSize))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktFlowWindow"}).Set(float64(stats.PktFlowWindow))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvBelated"}).Set(float64(stats.PktRcvBelated))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvBuf"}).Set(float64(stats.PktRcvBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvDrop"}).Set(float64(stats.PktRcvDrop))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvDropTotal"}).Set(float64(stats.PktRcvDropTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvFilterExtra"}).Set(float64(stats.PktRcvFilterExtra))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvFilterExtraTotal"}).Set(float64(stats.PktRcvFilterExtraTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvFilterLoss"}).Set(float64(stats.PktRcvFilterLoss))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvFilterLossTotal"}).Set(float64(stats.PktRcvFilterLossTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvFilterSupply"}).Set(float64(stats.PktRcvFilterSupply))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvFilterSupplyTotal"}).Set(float64(stats.PktRcvFilterSupplyTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvLoss"}).Set(float64(stats.PktRcvLoss))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvLossTotal"}).Set(float64(stats.PktRcvLossTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvRetrans"}).Set(float64(stats.PktRcvRetrans))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvUndecrypt"}).Set(float64(stats.PktRcvUndecrypt))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRcvUndecryptTotal"}).Set(float64(stats.PktRcvUndecryptTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRecv"}).Set(float64(stats.PktRecv))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRecvACK"}).Set(float64(stats.PktRecvACK))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRecvACKTotal"}).Set(float64(stats.PktRecvACKTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRecvNAK"}).Set(float64(stats.PktRecvNAK))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRecvNAKTotal"}).Set(float64(stats.PktRecvNAKTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRecvTotal"}).Set(float64(stats.PktRecvTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktReorderDistance"}).Set(float64(stats.PktReorderDistance))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktReorderTolerance"}).Set(float64(stats.PktReorderTolerance))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRetrans"}).Set(float64(stats.PktRetrans))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktRetransTotal"}).Set(float64(stats.PktRetransTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSent"}).Set(float64(stats.PktSent))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSentACK"}).Set(float64(stats.PktSentACK))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSentACKTotal"}).Set(float64(stats.PktSentACKTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSentNAK"}).Set(float64(stats.PktSentNAK))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSentNAKTotal"}).Set(float64(stats.PktSentNAKTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSentTotal"}).Set(float64(stats.PktSentTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSndBuf"}).Set(float64(stats.PktSndBuf))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSndDrop"}).Set(float64(stats.PktSndDrop))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSndDropTotal"}).Set(float64(stats.PktSndDropTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSndFilterExtra"}).Set(float64(stats.PktSndFilterExtra))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSndFilterExtraTotal"}).Set(float64(stats.PktSndFilterExtraTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSndLoss"}).Set(float64(stats.PktSndLoss))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "PktSndLossTotal"}).Set(float64(stats.PktSndLossTotal))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "UsSndDuration"}).Set(float64(stats.UsSndDuration))
	SrtStats.With(prometheus.Labels{"stream": streamid, "type": streamType, "remoteAddr": remoteAddr, "stat": "UsSndDurationTotal"}).Set(float64(stats.UsSndDurationTotal)) */
}
