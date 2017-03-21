package main

import (
	"testing"
)

func checkRedisCmd(t *testing.T, data []byte, expected string) {
	decoder := NewRedisDecoder("")
	result, err := decoder.Decode(data)
	if err != nil {
		t.Error(err)
	}
	assertEqual(t, result, expected)
}

func TestDecodeRedisMsgOK(t *testing.T) {
	checkRedisCmd(t, []byte("+OK\r\n"), "OK")
}

func TestDecodeRedisMsgERROR(t *testing.T) {
	checkRedisCmd(t, []byte("-ERROR\r\n"), "ERROR")
}

func TestDecodeRedisMsgInt(t *testing.T) {
	checkRedisCmd(t, []byte(":101\r\n"), "101")
}

func TestDecodeRedisMsgString(t *testing.T) {
	checkRedisCmd(t, []byte("$3\r\nget\r\n"), "get")
	checkRedisCmd(t, []byte("$-1\r\n"), "nil")
}

func TestDecodeRedisMsgArray(t *testing.T) {
	checkRedisCmd(t, []byte("*2\r\n$3\r\nget\r\n$1\r\na\r\n"), "get a")
}
