package main

import (
	"time"

	"github.com/RCVolus/srt-advanced-server/config"
	"github.com/RCVolus/srt-advanced-server/srt"
	"github.com/RCVolus/srt-advanced-server/web"
	"github.com/prometheus/client_golang/prometheus"
)

var Exit = make(chan string)

func main() {
	println("Starting SRT Advanced Server by RCVolus")

	go web.StartHttp()

	prometheus.MustRegister(srt.SrtStats)

	var config = config.GetConfig()

	time.Sleep(200 * time.Millisecond)

	for _, input := range config.Inputs {
		go srt.ListenIngressSocket(input.Port, input.Latency)
	}

	for _, output := range config.Outputs {
		go srt.ListenEgressSocket(output.Port, output.StreamId, output.Latency)
	}

	<-Exit
}
