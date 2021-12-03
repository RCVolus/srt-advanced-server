package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type StreamServerConfig struct {
	Inputs []struct {
		Port    uint16 `json:"port"`
		Latency uint16 `json:"latency"`
	} `json:"inputs"`
	Outputs []struct {
		StreamId string `json:"streamId"`
		Port     uint16 `json:"port"`
		Latency  uint16 `json:"latency"`
	} `json:"outputs"`
}

var config *StreamServerConfig

func loadConfig() {
	raw, err := ioutil.ReadFile("./cfg/config.json")
	if err != nil {
		log.Println("Error occured while reading config", err)
		return
	}
	json.Unmarshal(raw, &config)
}

func GetConfig() StreamServerConfig {
	if config == nil {
		loadConfig()
	}

	return *config
}
