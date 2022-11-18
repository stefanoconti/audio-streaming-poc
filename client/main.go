package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
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
	buffer := make([]float32, sampleRate*seconds)

	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, len(buffer), func(out []float32) {
		readFromServer(out, buffer, *server)
	})
	must(err)
	must(stream.Start())

	// clearTerminal()

	for {
		time.Sleep(time.Millisecond)
	}
}

func readFromServer(out []float32, buffer []float32, server string) {
	conn := dialServer(server)
	defer conn.Close()

	bs, _ := ioutil.ReadAll(conn)
	bytesReader := bytes.NewReader(bs)
	binary.Read(bytesReader, binary.BigEndian, &buffer)
	for i := range out {
		out[i] = buffer[i]
	}
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
