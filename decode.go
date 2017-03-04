package main

import (
	"fmt"
	"log"
)

func decodeHttp(data []byte) {
	log.Println("decode http")
}

func decodeRedis(data []byte) {
	log.Println("decode redis")
}

func decode(protocol string, data []byte) {
	if protocol == "ascii" {
		fmt.Println(string(data))
	} else if protocol == "http" {
		decodeHttp(data)
	} else if protocol == "redis" {
		decodeRedis(data)
	} else {
		panic("unknow protocol:" + protocol)
	}
}
