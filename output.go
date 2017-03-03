package main

import (
	"fmt"
	"net"
)

type StdoutOutput struct {
}

func (o *StdoutOutput) Write(data []byte) {
	fmt.Println(string(data))
}

func (o *StdoutOutput) Close() {
}

type TCPOutput struct {
	conn net.Conn
}

func NewTcpOutput(addr string) (*TCPOutput, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPOutput{conn}, _
}

func (o *TCPOutput) Write(data []byte) {
	o.conn.Write(data)
}

func (o *TCPOutput) Close() {
	o.conn.Close()
}
