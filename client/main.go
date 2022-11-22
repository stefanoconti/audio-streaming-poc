package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/gordonklaus/portaudio"
)

const sampleRate = 11025
const seconds = 0.04

func main() {

	server := flag.String("server", "localhost:8080", "the server to connect to")

	flag.Parse()

	portaudio.Initialize()
	defer portaudio.Terminate()

	conn := dialServer(*server)
	defer conn.Close()

	// chin := make(chan []byte)
	chout := make(chan []float32, 2)

	// go readConn(chin, conn)
	// go processStream(chin, chout)

	go processIncomingData(conn, chout)

	//buffer := make([]float32, sampleRate*seconds)

	h, err := portaudio.DefaultHostApi()
	must(err)

	for i, d := range h.Devices {
		fmt.Println(i, d.Name, d.MaxInputChannels, d.MaxOutputChannels)
	}

	p := portaudio.LowLatencyParameters(nil, h.Devices[2])
	p.Input.Channels = 0
	p.Output.Channels = 1
	p.SampleRate = sampleRate

	stream, err := portaudio.OpenStream(p, func(out []float32) {
		//readFromServer2(out, buffer, conn)
		writeAudio(chout, out)
	})

	/*
		stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, len(buffer), func(out []float32) {
			//readFromServer2(out, buffer, conn)
			writeAudio(chout, out)
		})
	*/

	must(err)
	must(stream.Start())

	// clearTerminal()

	/*
		for {
			time.Sleep(time.Millisecond)
		}
	*/

	keepAlive := make(chan os.Signal)
	exitStatus := 0
	signal.Notify(keepAlive, os.Kill, os.Interrupt)

	<-keepAlive
	fmt.Println("Bye!")
	os.Exit(exitStatus)
}

/*
func readFromServerProfiling(out []float32, buffer []float32, conn net.Conn) {
	start := time.Now().UnixMicro()
	readFromServer2(out, buffer, conn)
	stop := time.Now().UnixMicro() - start
	fmt.Println(stop)
}
*/
/*
func readFromServer(out []float32, buffer []float32, conn net.Conn) {
	bs, _ := ioutil.ReadAll(conn)
	bytesReader := bytes.NewReader(bs)
	binary.Read(bytesReader, binary.BigEndian, &buffer)
	for i := range out {
		out[i] = buffer[i]
	}
}
*/

func processIncomingData(conn net.Conn, out chan []float32) {
	for {
		bs, _ := ioutil.ReadAll(conn)
		bytesReader := bytes.NewReader(bs)
		outBuf := make([]float32, sampleRate*seconds)
		binary.Read(bytesReader, binary.BigEndian, &outBuf)
		out <- outBuf
	}
}

func readConn(ch chan []byte, conn net.Conn) {
	for {
		bs, _ := ioutil.ReadAll(conn)
		ch <- bs
	}
}

func writeAudio(ch chan []float32, out []float32) {
	bs := <-ch
	copy(out, bs)
}

func processStream(in chan []byte, out chan []float32) {
	bs := <-in
	bytesReader := bytes.NewReader(bs)
	outBuf := make([]float32, sampleRate*seconds)
	binary.Read(bytesReader, binary.BigEndian, &outBuf)
	out <- outBuf
}

/*
func readFromServer2(out []float32, buffer []float32, conn net.Conn) {
	bs, _ := ioutil.ReadAll(conn)
	bytesReader := bytes.NewReader(bs)
	bufReader := bufio.NewReaderSize(bytesReader, len(buffer)+1)
	binary.Read(bufReader, binary.BigEndian, &out)
}
*/

func dialServer(server string) net.Conn {
	conn, errConn := net.Dial("tcp", server)
	for errConn != nil {
		conn, errConn = net.Dial("tcp", server)
		fmt.Println("Trying to reconnect...")
		time.Sleep(time.Second)
	}
	return conn
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
