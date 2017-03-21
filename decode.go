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
		return new(AsciiDecoder), nil
	case "http":
		return NewHttpDecoder(filterStr), nil
	case "redis":
		return NewRedisDecoder(filterStr), nil
	default:
		return nil, errors.New("unknow protocol: " + decodeAs)
	}
}
