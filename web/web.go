package web

import (
	"net/http"
	"os"

	"github.com/RCVolus/srt-advanced-server/stream"
	"github.com/haivision/srtgo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type OutputInfo struct {
	StreamInfo  stream.EgestStreamInformation
	StreamStats srtgo.SrtStats
}

type StreamInfo struct {
	StreamInfo  stream.IngestStreamInformation
	StreamStats srtgo.SrtStats
	Outputs     []OutputInfo
}

func StartHttp() {
	e := echo.New()

	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		streamInfo := make(map[string]StreamInfo)

		for key, value := range stream.IngestStreams {
			var outputs []OutputInfo

			for _, output := range value.Outputs {
				stats, _ := output.Socket.Stats()
				outputs = append(outputs, OutputInfo{
					StreamInfo:  output.EgestStreamInformation,
					StreamStats: *stats,
				})
			}

			stats, _ := value.Socket.Stats()
			streamInfo[key] = StreamInfo{
				StreamInfo:  value.IngestStreamInformation,
				StreamStats: *stats,
				Outputs:     outputs,
			}
		}

		return c.JSON(http.StatusOK, streamInfo)
	})

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "11000"
	}

	e.Start(":" + httpPort)
}
