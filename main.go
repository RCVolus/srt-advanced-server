package main

import (
	"github.com/RCVolus/srt-advanced-server/srt"
	"github.com/RCVolus/srt-advanced-server/web"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	go web.StartHttp()

	prometheus.MustRegister(srt.SrtStats)

	go srt.ListenEgressSocket()
	srt.ListenIngressSocket()
}
