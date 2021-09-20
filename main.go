package main

import (
	"github.com/RCVolus/srt-advanced-server/srt"
	"github.com/RCVolus/srt-advanced-server/web"
)

func main() {
	go web.StartHttp()

	go srt.ListenEgressSocket()
	srt.ListenIngressSocket()
}
