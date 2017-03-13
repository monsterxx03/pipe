package main

import "bytes"

type httpDecoder struct {
	buf bytes.Buffer
}

func (d *httpDecoder) decode() (string, error) {
	return "http" + d.buf.String(), nil
}

func newHttpDecoder(data []byte) *httpDecoder {
	return &httpDecoder{*bytes.NewBuffer(data)}
}
