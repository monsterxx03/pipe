package main

import (
	"testing"
)

func TestDecodeHttpReq(t *testing.T) {
	decoder := NewHttpDecoder("")
	data := []byte("POST /test HTTP/1.1\r\nHost: google.com\r\nUser-Agent:curl\r\n\r\nHello\r\n")
	decoder.write(data)
	msg, err := decoder.decodeHttp()
	assertEqual(t, err, nil)
	req := msg.(*HttpReq)
	assertEqual(t, req.method, "POST")
	assertEqual(t, req.url, "/test")
	assertEqual(t, req.headers["Host"], "google.com")
	assertEqual(t, req.headers["User-Agent"], "curl")
	assertEqual(t, string(req.body), "Hello\r\n")
}

func TestDecodeHttpResp(t *testing.T) {
	decoder := NewHttpDecoder("")
	data := []byte("HTTP/1.1 200 OK\r\nHost: google.com\r\n\r\nHello World")
	decoder.write(data)
	msg, err := decoder.decodeHttp()
	assertEqual(t, err, nil)
	resp := msg.(*HttpResp)
	assertEqual(t, resp.statusCode, 200)
	assertEqual(t, resp.statusMsg, "OK")
	assertEqual(t, resp.headers["Host"], "google.com")
	assertEqual(t, string(resp.body), "Hello World")
}

func TestHttpReqFilterUrl(t *testing.T) {
	decoder := NewHttpDecoder("url: /test")
	data := []byte("POST /tes/haha HTTP/1.1\r\nHost: google.com\r\nUser-Agent:curl\r\n\r\nHello\r\n")
	decoder.write(data)
	// url not match
	msg, err := decoder.decodeHttp()
	assertEqual(t, msg, nil)
	assertEqual(t, err, SkipError)
}

func TestHttpRespFilter(t *testing.T) {

}
