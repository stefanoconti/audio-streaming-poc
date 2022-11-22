package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/gordonklaus/portaudio"
	"github.com/stefanoconti/audio-streaming-poc/internal/common"
)

const sampleRate = 11025
const seconds = 0.04
const frames = sampleRate * seconds

func main() {

	address := flag.String("address", ":8080", "the address to bind to")

	flag.Parse()

	portaudio.Initialize()
	defer portaudio.Terminate()

	channelAudioInput := make(chan []float32)
	channelAudioOutput := make(chan []float32)

	h, err := portaudio.DefaultHostApi()
	must(err)
	for i, d := range h.Devices {
		fmt.Println(i, d.Name, d.MaxInputChannels, d.MaxOutputChannels)
	}

	pIn := portaudio.LowLatencyParameters(h.DefaultInputDevice, nil)
	pIn.Input.Channels = 1
	pIn.Output.Channels = 0
	pIn.SampleRate = sampleRate
	pIn.FramesPerBuffer = frames

	streamIn, err := portaudio.OpenStream(pIn, func(in []float32) {
		common.ReadAudioStream(in, frames, channelAudioInput)
	})

	must(err)
	must(streamIn.Start())
	defer streamIn.Close()

	pOut := portaudio.LowLatencyParameters(nil, h.DefaultOutputDevice)
	pIn.Input.Channels = 0
	pIn.Output.Channels = 1
	pIn.SampleRate = sampleRate
	pIn.FramesPerBuffer = frames

	streamOut, err := portaudio.OpenStream(pOut, func(out []float32) {
		common.WriteAudioStream(out, channelAudioOutput)
	})

	must(err)
	must(streamOut.Start())
	defer streamOut.Close()

	listen, errNet := net.Listen("tcp", *address)
	if errNet != nil {
		log.Fatal(errNet)
	}
	defer listen.Close()

	fmt.Println("Running Server TCP on port ", *address)

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go common.WriteOutcomingData(conn, channelAudioInput)
		go common.ReadIncomingData(conn, frames, channelAudioOutput)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
