package main

import (
	"errors"
	_ "log"
)

type Decoder interface {
	decode() (string, error)
}

type asciiDecoder struct {
	data []byte
}

func (d *asciiDecoder) decode() (string, error) {
	return string(d.data), nil
}

func decode(protocol string, data []byte) (string, error) {
	var d Decoder
	switch protocol {
	case "ascii":
		d = &asciiDecoder{data}
	case "http":
		d = newHttpDecoder(data)
	case "redis":
		d = newRedisDecoder(data)
	default:
		return "", errors.New("unknown protocol: " + protocol)
	}
	if resp, err := d.decode(); err != nil {
		return "", err
	} else {
		return resp, nil
	}
}
