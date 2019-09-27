package pkg

import (
	"github.com/pion/sctp"
	"io"
	"log"
)

type FlowControlledStream struct {
	stream            *sctp.Stream
	maxBufferedAmount uint64
	queue             chan []byte
}

// NewFlowControlledStream --
func NewFlowControlledStream(stream *sctp.Stream, bufferedAmountLowThreshold, maxBufferedAmount, queueSize uint64) io.ReadWriter {
	stream.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)
	fcs := &FlowControlledStream{
		stream:            stream,
		maxBufferedAmount: maxBufferedAmount,
		queue:             make(chan []byte, queueSize),
	}

	stream.OnBufferedAmountLow(func() {
		go fcs.DrainQueue()
	})

	return fcs
}

func (fcdc *FlowControlledStream) Read(p []byte) (int, error) {
	return fcdc.stream.Read(p)
}

func (fcdc *FlowControlledStream) Write(p []byte) (int, error) {
	fcdc.queue <- p
	if _, err := fcdc.DrainQueue(); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (fcdc *FlowControlledStream) DrainQueue() (int, error) {
	var bytesSent int
	for {
		if len(fcdc.queue) == 0 {
			break
		}

		if fcdc.stream.BufferedAmount() >= fcdc.maxBufferedAmount {
			break
		}

		p := <-fcdc.queue
		b, err := fcdc.stream.Write(p)
		if err != nil {
			log.Println("ERROR", err)
			return bytesSent, err
		}

		bytesSent += b
	}
	return bytesSent, nil
}
