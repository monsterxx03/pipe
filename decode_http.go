package main

import (
	"bytes"
	"log"
	"strings"
)

const (
	httpParseMethod = iota
	httpParseHeader
	httpParseBody
	httpParseDone
)

type HttpDecoder struct {
	buf    bytes.Buffer
	filter *HttpFilter
}

type HttpMsg struct {
	method  string
	url     string
	version string
	headers map[string]string
	body    []byte
}

func (m *HttpMsg) String() string {
	// TODO format output
	return m.url
}

func (d *HttpDecoder) Decode(data []byte) (string, error) {
	if len(data) > 0 {
		d.buf.Write(data)
	}
	msg := new(HttpMsg)
	msg.headers = make(map[string]string)
	var line string
	var err error
	var currentPosition int = httpParseMethod
	for {
		log.Println("loop")
		switch currentPosition {
		case httpParseMethod:
			// parse first line
			line, err = d.buf.ReadString('\n')
			if err != nil {
				return "", err
			}
			line = line[:len(line)-2] // remove \r\n
			result := strings.Split(line, " ")
			msg.method = strings.TrimSpace(result[0])
			msg.url = strings.TrimSpace(result[1])
			msg.version = strings.TrimSpace(result[2])
			currentPosition = httpParseHeader
		case httpParseHeader:
			// parse headers
			line, err = d.buf.ReadString('\n')
			if err != nil {
				return "", err
			}
			if line == "\r\n" {
				currentPosition = httpParseBody
				break
			}
			line = line[:len(line)-2] // remove \r\n
			result := strings.Split(line, " ")
			result = strings.SplitN(line, ":", 2)
			msg.headers[result[0]] = strings.TrimSpace(result[1])
		case httpParseBody:
			// get body
			msg.body = make([]byte, len(d.buf.Bytes()))
			copy(msg.body, d.buf.Bytes())
			d.buf.Reset()
			currentPosition = httpParseDone
		}
		if currentPosition == httpParseDone {
			break
		}
	}
	if !d.filter.IsEmpty() {
		// TODO filter msg
	}
	return msg.String(), nil
}

func NewHttpDecoder(filterStr string) *HttpDecoder {
	return &HttpDecoder{filter: NewHttpFilter(filterStr)}
}
