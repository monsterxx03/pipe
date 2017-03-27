package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type HttpMsg interface {
	Match(f *HttpFilter) bool
}

type HttpMixin struct {
	headers map[string]string
	body    []byte
	rawMsg  []byte
}

func (m *HttpMixin) parseHeader(buf *bytes.Buffer) error {
	var err error
	var line string
	for {
		// parse headers
		if line, err = buf.ReadString('\n'); err != nil {
			return err
		}
		if line == "\r\n" {
			// end of header
			break
		}
		line = line[:len(line)-2] // remove \r\n
		result := strings.SplitN(line, ":", 2)
		m.headers[result[0]] = strings.TrimSpace(result[1])
	}
	return nil
}

func (m *HttpMixin) parseBody(buf *bytes.Buffer) error {
	m.body = make([]byte, len(buf.Bytes()))
	copy(m.body, buf.Bytes())
	buf.Reset()
	return nil
}

func (m HttpMixin) String() string {
	return string(m.rawMsg)
}

func matchHeaders(rules map[string]*regexp.Regexp, headers map[string]string) bool {
	for h, pattern := range rules {
		value, ok := headers[h]
		if !ok {
			// header not exist, so not match
			return false
		}
		if !pattern.MatchString(value) {
			return false
		}
	}
	return true
}

func matchBody(pattern *regexp.Regexp, body []byte) bool {
	if !pattern.Match(body) {
		return false
	}
	return true
}

type HttpReq struct {
	method  string
	url     string
	version string
	HttpMixin
}

func (m *HttpReq) Match(filter *HttpFilter) bool {
	filters := make(map[string]*regexp.Regexp)
	MapCopy(filters, filter.filters)
	if _, ok := filters["method"]; ok && !filters["method"].MatchString(m.method) {
		return false
	}
	delete(filters, "method")
	if _, ok := filters["url"]; ok && !filters["url"].MatchString(m.url) {
		return false
	}
	delete(filters, "url")
	if _, ok := filters["version"]; ok && !filters["version"].MatchString(m.version) {
		return false
	}
	delete(filters, "version")
	if _, ok := filters["body"]; ok && !matchBody(filters["body"], m.body) {
		return false
	}
	delete(filters, "body")
	if len(filters) > 0 && !matchHeaders(filters, m.headers) {
		return false
	}
	return true
}

type HttpResp struct {
	version    string
	statusCode int
	statusMsg  string
	HttpMixin
}

func (m *HttpResp) Match(filter *HttpFilter) bool {
	filters := make(map[string]*regexp.Regexp)
	MapCopy(filters, filter.filters)
	if _, ok := filters["version"]; ok && !filters["version"].MatchString(m.version) {
		return false
	}
	delete(filters, "version")
	if _, ok := filters["statusCode"]; ok && !filters["statusCode"].MatchString(strconv.Itoa(m.statusCode)) {
		return false
	}
	delete(filters, "statusCode")
	if _, ok := filters["statusMsg"]; ok && !filters["statusMsg"].MatchString(m.statusMsg) {
		return false
	}
	delete(filters, "statusMsg")
	if _, ok := filters["body"]; ok && !matchBody(filters["body"], m.body) {
		return false
	}
	delete(filters, "body")
	if len(filters) > 0 && !matchHeaders(filters, m.headers) {
		return false
	}
	return true
}

type HttpDecoder struct {
	buf    *bytes.Buffer
	filter *HttpFilter
}

func (d *HttpDecoder) Decode(data []byte) (string, error) {
	d.write(data)
	if msg, err := d.decodeHttp(); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("%v", msg), nil
	}
}

func (d *HttpDecoder) write(data []byte) {
	if len(data) > 0 {
		d.buf.Write(data)
	}
}

func (d *HttpDecoder) decodeHttp() (HttpMsg, error) {
	var err error
	var msg HttpMsg
	if d.buf.Bytes()[0] == 72 && d.buf.Bytes()[1] == 84 { // starts with HT
		msg, err = d.parseResponse()
	} else {
		msg, err = d.parseRequest()
	}
	if err != nil {
		return nil, err
	}

	if !d.filter.IsEmpty() && !msg.Match(d.filter) {
		return nil, SkipError
	}
	return msg, nil
}

func (d *HttpDecoder) parseRequest() (*HttpReq, error) {
	var line string
	var err error
	msg := new(HttpReq)
	msg.rawMsg = make([]byte, len(d.buf.Bytes()))
	copy(msg.rawMsg, d.buf.Bytes())
	line, err = d.buf.ReadString('\n')
	if err != nil {
		return nil, err
	}
	// parse first line
	line = line[:len(line)-2] // remove \r\n
	result := strings.Split(line, " ")
	msg.method = strings.TrimSpace(result[0])
	msg.url = strings.TrimSpace(result[1])
	msg.version = strings.TrimSpace(result[2])
	msg.headers = make(map[string]string)
	if err = msg.parseHeader(d.buf); err != nil {
		return nil, err
	}
	if err = msg.parseBody(d.buf); err != nil {
		return nil, err
	}
	return msg, nil
}

func (d *HttpDecoder) parseResponse() (*HttpResp, error) {
	var line string
	var err error
	msg := new(HttpResp)
	msg.rawMsg = make([]byte, len(d.buf.Bytes()))
	copy(msg.rawMsg, d.buf.Bytes())
	line, err = d.buf.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = line[:len(line)-2] // remove \r\n
	result := strings.Split(line, " ")
	msg.version = strings.TrimSpace(result[0])
	msg.statusCode, err = strconv.Atoi(strings.TrimSpace(result[1]))
	if err != nil {
		return nil, err
	}
	msg.statusMsg = strings.TrimSpace(result[2])
	msg.headers = make(map[string]string)
	if err = msg.parseHeader(d.buf); err != nil {
		return nil, err
	}
	if err = msg.parseBody(d.buf); err != nil {
		return nil, err
	}
	return msg, nil
}

func NewHttpDecoder(filterStr string) *HttpDecoder {
	return &HttpDecoder{buf: bytes.NewBuffer([]byte{}),
		filter: NewHttpFilter(filterStr)}
}
