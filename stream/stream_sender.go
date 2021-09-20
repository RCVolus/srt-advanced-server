package stream

import (
	"sync"
	"time"

	"github.com/haivision/srtgo"
)

var IngestStreams = make(map[string]StreamSender)

type StreamReceiver struct {
	EgestStreamInformation EgestStreamInformation
	Socket                 srtgo.SrtSocket
}

type StreamSender struct {
	Broadcast               chan<- []byte
	IngestStreamInformation IngestStreamInformation
	Socket                  srtgo.SrtSocket

	Outputs map[chan []byte]StreamReceiver

	LockOutputs sync.Mutex
}

type EgestStreamInformation struct {
	Remote      string
	ConnectedAt time.Time
}

type IngestStreamInformation struct {
	StreamId    string
	Remote      string
	ConnectedAt time.Time
}

func NewStreamSender(ingestStreamInformation IngestStreamInformation, socket srtgo.SrtSocket) (streamSender *StreamSender) {
	streamSender = &StreamSender{}
	broadcast := make(chan []byte, 1024)
	streamSender.Broadcast = broadcast
	streamSender.Outputs = make(map[chan []byte]StreamReceiver)
	streamSender.IngestStreamInformation = ingestStreamInformation
	streamSender.Socket = socket

	go streamSender.run(broadcast)
	return streamSender
}

func (streamSender *StreamSender) run(broadcast <-chan []byte) {
	for msg := range broadcast {
		streamSender.LockOutputs.Lock()
		for output := range streamSender.Outputs {
			select {
			case output <- msg:
			default:
				if len(output) > 1 {
					<-output
				}
			}
		}
		streamSender.LockOutputs.Unlock()
	}

	// The sender channel has been closed, close all outputs
	streamSender.LockOutputs.Lock()
	for ch := range streamSender.Outputs {
		delete(streamSender.Outputs, ch)
		close(ch)
	}
	streamSender.LockOutputs.Unlock()
}

// Close stream sender and remove all outputs
func (streamSender *StreamSender) Close() {
	close(streamSender.Broadcast)
}

// Register a new output
func (streamSender *StreamSender) Register(output chan []byte, info EgestStreamInformation, socket srtgo.SrtSocket) {
	streamSender.LockOutputs.Lock()
	streamSender.Outputs[output] = StreamReceiver{
		EgestStreamInformation: info,
		Socket:                 socket,
	}
	streamSender.LockOutputs.Unlock()
}

// Remove an output
func (streamSender *StreamSender) Unregister(output chan []byte) {
	streamSender.LockOutputs.Lock()
	_, ok := streamSender.Outputs[output]
	if ok {
		delete(streamSender.Outputs, output)
		close(output)
	}
	defer streamSender.LockOutputs.Unlock()
}
