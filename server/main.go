package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/gordonklaus/portaudio"
)

const sampleRate = 11025
const seconds = 0.04

func main() {

	address := flag.String("address", ":8080", "the address to bind to")

	flag.Parse()

	portaudio.Initialize()
	defer portaudio.Terminate()

	chin := make(chan []float32, 2)

	//buffer := make([]float32, sampleRate*seconds)

	h, err := portaudio.DefaultHostApi()
	must(err)

	for i, d := range h.Devices {
		fmt.Println(i, d.Name, d.MaxInputChannels, d.MaxOutputChannels)
	}

	p := portaudio.LowLatencyParameters(h.Devices[3], nil)
	p.Input.Channels = 1
	p.Output.Channels = 0
	p.SampleRate = sampleRate

	stream, err := portaudio.OpenStream(p, func(in []float32) {
		buf := make([]float32, sampleRate*seconds)
		copy(buf, in)
		chin <- buf

	})

	/*
		stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, len(buffer), func(in []float32) {
			buf := make([]float32, sampleRate*seconds)
			copy(buf, in)
			chin <- buf
		})
	*/

	must(err)
	must(stream.Start())
	defer stream.Close()

	listen, errNet := net.Listen("tcp", *address)
	if errNet != nil {
		log.Fatal(errNet)
	}
	defer listen.Close()

	// clearTerminal()
	fmt.Println("Running Server TCP on port ", *address)

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		// go handle(conn, buffer)
		go processOutcomingData(conn, chin)
	}
}

func handle(con net.Conn, buffer []float32) {
	defer con.Close()
	binary.Write(con, binary.BigEndian, &buffer)
}

func processOutcomingData(conn net.Conn, in chan []float32) {
	defer conn.Close()
	for {
		buf := <-in
		binary.Write(conn, binary.BigEndian, &buf)
	}
}

/*
func clearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
*/

func must(err error) {
	if err != nil {
		panic(err)
	}
}
