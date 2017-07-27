package redis

import (
	"testing"
	"bytes"
	"io"
	"bufio"
)

func checkRedisCmd(t *testing.T, data []byte, expected string) {
	decoder := Decoder{}
	decoder.SetFilter("")
	pr, pw := io.Pipe()
	go func() {
		err := decoder.Decode(bytes.NewReader(data), pw)
		if err != nil && err != io.EOF {
			t.Error(err)
		}
	}()

	bufPr := bufio.NewReader(pr)
	_data, err := bufPr.ReadString('\n')
	if err != nil {
		t.Error(err)
	}
	_data = _data[:len(_data)-1]
	if string(_data) != expected {
		t.Error("not equal")
	}
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
