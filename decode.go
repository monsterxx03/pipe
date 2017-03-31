package main

import (
	"bufio"
	"errors"
	_ "log"
)

var SkipError = errors.New("skip packet")

type Decoder interface {
	Decode() (string, error)
	SetReader(r *bufio.Reader)
}

type AsciiDecoder struct {
	r *bufio.Reader
}

func (d *AsciiDecoder) Decode() (string, error) {
	return "Test", nil
}

func (d *AsciiDecoder) SetReader(r *bufio.Reader) {
	d.r = r
}

func getDecoder(decodeAs, filterStr string) (Decoder, error) {
	switch decodeAs {
	case "ascii":
		return new(AsciiDecoder), nil
	case "http":
		return NewHttpDecoder(filterStr), nil
	case "redis":
		return NewRedisDecoder(filterStr), nil
	default:
		return nil, errors.New("unknow protocol: " + decodeAs)
	}
}
