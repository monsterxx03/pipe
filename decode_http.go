package main

import "bytes"

type HttpDecoder struct {
	buf    bytes.Buffer
	filter *HttpFilter
}

type HttpMsg struct {
	method  string
	url     string
	headers map[string]string
	body    string
}

func (d *HttpDecoder) Decode(data []byte) (string, error) {
	d.buf.Write(data)
	return d.buf.String(), nil
}

func NewHttpDecoder(filterStr string) *HttpDecoder {
	return &HttpDecoder{filter: NewHttpFilter(filterStr)}
}
