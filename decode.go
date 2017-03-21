package main

import (
	"errors"
	_ "log"
)

type Decoder interface {
	Decode([]byte) (string, error)
}

type AsciiDecoder struct {
	data []byte
}

func (d *AsciiDecoder) Decode(data []byte) (string, error) {
	return string(data), nil
}

func GetDecoder(decodeAs, filterStr string) (Decoder, error) {
	switch decodeAs {
	case "ascii":
		d := new(AsciiDecoder)
		return d, nil
	case "http":
		d := NewHttpDecoder(filterStr)
		return d, nil
	case "redis":
		d := NewRedisDecoder(filterStr)
		return d, nil
	default:
		return nil, errors.New("unknow protocol: " + decodeAs)
	}
}
