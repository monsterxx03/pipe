package main

import (
	"errors"
	"fmt"
	_ "log"
)

type decodeFunc func(data []byte) (interface{}, error)

var decodeMap = map[string]decodeFunc{
	"ascii": func(data []byte) (interface{}, error) {
		fmt.Println(string(data))
		return nil, nil
	},
	"http":  decodeHttp,
	"redis": decodeRedis,
}

func decode(protocol string, data []byte) error {
	decoder, ok := decodeMap[protocol]
	if ok {
		decoder(data)
		return nil
	} else {
		return errors.New("unknown protocol: " + protocol)
	}
}
