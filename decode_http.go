package main

import (
	"bytes"
	"errors"
	"log"
	"strconv"
	"strings"
)

const (
	httpParseFirstLine = iota
	httpParseHeader
	httpParseBody
	httpParseDone
)

type HttpDecoder struct {
	buf    *bytes.Buffer
	filter *HttpFilter
}

type HttpMixin struct {
	headers map[string]string
	body    []byte
}

type HttpMsg interface {
	parseHeader(*bytes.Buffer) error
	parseBody(*bytes.Buffer) error
}

func (m *HttpMixin) parseHeader(buf *bytes.Buffer) error {
	var err error
	var line string
	for {
		// parse headers
		if line, err = buf.ReadString('\n'); err != nil {
			return err
		}
		if line == "\r\n" {
			// end of header
			break
		}
		line = line[:len(line)-2] // remove \r\n
		result := strings.SplitN(line, ":", 2)
		m.headers[result[0]] = strings.TrimSpace(result[1])
	}
	return nil
}

func (m *HttpMixin) parseBody(buf *bytes.Buffer) error {
	m.body = make([]byte, len(buf.Bytes()))
	copy(m.body, buf.Bytes())
	buf.Reset()
	return nil
}

type HttpReq struct {
	method  string
	url     string
	version string
	HttpMixin
}

type HttpResp struct {
	version    string
	statusCode int
	statusMsg  string
	HttpMixin
}

func (m *HttpReq) String() string {
	// TODO format output
	return m.url
}

func (m *HttpResp) String() string {
	// TODO format output
	return m.statusMsg
}

func (d *HttpDecoder) Decode(data []byte) (string, error) {
	if len(data) > 0 {
		d.buf.Write(data)
	}
	var err error
	var msg HttpMsg
	if data[0] == 72 && data[1] == 84 { // starts with HT
		msg = new(HttpResp)
		err = d.parseResponse(msg.(*HttpResp))
	} else {
		msg = new(HttpReq)
		err = d.parseRequest(msg.(*HttpReq))
	}
	if err = msg.parseHeader(d.buf); err != nil {
		return "", err
	}
	if err = msg.parseBody(d.buf); err != nil {
		return "", err
	}
	log.Println(msg)

	if !d.filter.IsEmpty() && d.filter.Match(msg) {
		return "", errors.New("skip packet")
	}
	return "", nil
}

func (d *HttpDecoder) parseRequest(msg *HttpReq) error {
	var line string
	var err error
	line, err = d.buf.ReadString('\n')
	if err != nil {
		return err
	}
	// parse first line
	line = line[:len(line)-2] // remove \r\n
	result := strings.Split(line, " ")
	msg.method = strings.TrimSpace(result[0])
	msg.url = strings.TrimSpace(result[1])
	msg.version = strings.TrimSpace(result[2])
	msg.headers = make(map[string]string)
	return nil
}

func (d *HttpDecoder) parseResponse(msg *HttpResp) error {
	var line string
	var err error
	line, err = d.buf.ReadString('\n')
	if err != nil {
		return err
	}
	line = line[:len(line)-2] // remove \r\n
	result := strings.Split(line, " ")
	msg.version = strings.TrimSpace(result[0])
	msg.statusCode, err = strconv.Atoi(strings.TrimSpace(result[1]))
	if err != nil {
		return err
	}
	msg.statusMsg = strings.TrimSpace(result[2])
	msg.headers = make(map[string]string)
	return nil
}

func NewHttpDecoder(filterStr string) *HttpDecoder {
	return &HttpDecoder{buf: bytes.NewBuffer([]byte{}), filter: NewHttpFilter(filterStr)}
}
