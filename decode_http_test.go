package main

import (
	"bufio"
	"bytes"
	"testing"
)

func TestDecodeHttpReq(t *testing.T) {
	decoder := NewHttpDecoder("")
	data := []byte("POST /test HTTP/1.1\r\nHost: google.com\r\nUser-Agent:curl\r\nContent-Length: 5\r\n\r\nHello")
	r := bufio.NewReader(bytes.NewReader(data))
	decoder.SetReader(r)
	msg, err := decoder.decodeHttp()
	assertEqual(t, err, nil)
	req := msg.(*HttpReq)
	assertEqual(t, req.method, "POST")
	assertEqual(t, req.url, "/test")
	assertEqual(t, req.headers["Host"], "google.com")
	assertEqual(t, req.headers["User-Agent"], "curl")
	assertEqual(t, string(req.body), "Hello")
}

func TestDecodeHttpResp(t *testing.T) {
	decoder := NewHttpDecoder("")
	data := []byte("HTTP/1.1 200 OK\r\nContent-Length:11\r\nHost: google.com\r\n\r\nHello World")
	r := bufio.NewReader(bytes.NewReader(data))
	decoder.SetReader(r)
	msg, err := decoder.decodeHttp()
	assertEqual(t, err, nil)
	resp := msg.(*HttpResp)
	assertEqual(t, resp.statusCode, 200)
	assertEqual(t, resp.statusMsg, "OK")
	assertEqual(t, resp.headers["Host"], "google.com")
	assertEqual(t, string(resp.body), "Hello World")
}

func TestHttpReqFilter(t *testing.T) {
	decoder := NewHttpDecoder("url: /test & method: POST")
	data := []byte("POST /tes/haha HTTP/1.1\r\nHost: google.com\r\nUser-Agent:curl\r\n\r\nHello\r\n")
	r := bufio.NewReader(bytes.NewReader(data))
	decoder.SetReader(r)
	// url not match
	msg, err := decoder.decodeHttp()
	assertEqual(t, msg, nil)
	assertEqual(t, err, SkipError)

	// match
	data = []byte("POST /test/haha HTTP/1.1\r\nHost: google.com\r\nUser-Agent:curl\r\n\r\nHello\r\n")
	r = bufio.NewReader(bytes.NewReader(data))
	decoder.SetReader(r)
	msg, err = decoder.decodeHttp()
	assertEqual(t, err, nil)
}

func TestHttpRespFilter(t *testing.T) {

}
