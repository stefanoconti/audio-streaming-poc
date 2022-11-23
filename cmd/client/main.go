package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/stefanoconti/audio-streaming-poc/internal/common"
)

const sampleRate = 11025
const seconds = 0.04
const frames = sampleRate * seconds

func main() {

	server := flag.String("server", "localhost:8080", "the server to connect to")

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
	pOut.Input.Channels = 0
	pOut.Output.Channels = 1
	pOut.SampleRate = sampleRate
	pOut.FramesPerBuffer = frames

	streamOut, err := portaudio.OpenStream(pOut, func(out []float32) {
		common.WriteAudioStream(out, channelAudioOutput)
	})

	must(err)
	must(streamOut.Start())
	defer streamOut.Close()

	conn := dialServer(*server)
	defer conn.Close()

	keepAlive := make(chan os.Signal)
	exitStatus := 0
	signal.Notify(keepAlive, os.Kill, os.Interrupt)

	<-keepAlive
	fmt.Println("Bye!")
	os.Exit(exitStatus)
}

func dialServer(server string) net.Conn {
	conn, errConn := net.Dial("tcp", server)
	for errConn != nil {
		conn, errConn = net.Dial("tcp", server)
		fmt.Println("Trying to reconnect...")
		time.Sleep(time.Second)
	}
	return conn
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
