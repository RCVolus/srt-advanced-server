package web

import (
	"fmt"
	"net/http"
	"os"

	"github.com/RCVolus/srt-advanced-server/stream"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type OutputInfo struct {
	StreamInfo stream.EgestStreamInformation
}

type StreamInfo struct {
	StreamInfo stream.IngestStreamInformation
	Outputs    OutputInfo
}

func StartHttp() {
	e := echo.New()

	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, fmt.Sprintf("Welcome to srt-advanced-server, ingest streams: %d", len(stream.IngestStreams)))
	})

	e.GET("/ping", func(c echo.Context) error {
		outputs := make(map[string]StreamInfo)

		for key, value := range stream.IngestStreams {
			outputs[key] = StreamInfo{
				StreamInfo: value.IngestStreamInformation,
				NumOutputs: len(value.Outputs),
			}
		}

		return c.JSON(http.StatusOK, outputs)
	})

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "11000"
	}

	e.Start(":" + httpPort)
}
