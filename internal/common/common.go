package common

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"net"
)

func ReadIncomingData(conn net.Conn, framesPerBuffer int, audioOut chan []float32) {
	for {
		bs, _ := ioutil.ReadAll(conn)
		bytesReader := bytes.NewReader(bs)
		buffer := make([]float32, framesPerBuffer)
		binary.Read(bytesReader, binary.BigEndian, &buffer)
		audioOut <- buffer
	}
}

func WriteOutcomingData(conn net.Conn, audioIn chan []float32) {
	for {
		buffer := <-audioIn
		binary.Write(conn, binary.BigEndian, buffer)
	}
}

func ReadAudioStream(in []float32, framesPerBuffer int, audioIn chan []float32) {
	buffer := make([]float32, framesPerBuffer)
	copy(buffer, in)
	audioIn <- buffer
}

func WriteAudioStream(out []float32, audioOut chan []float32) {
	buffer := <-audioOut
	copy(out, buffer)
}
