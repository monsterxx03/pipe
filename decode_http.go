package main

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type HttpMsg interface{}

type HttpMixin struct {
	headers map[string]string
	body    []byte
}

func (m *HttpMixin) parseHeader(reader *bufio.Reader) error {
	var err error
	var line string
	m.headers = make(map[string]string)
	for {
		// parse headers
		if line, err = reader.ReadString('\n'); err != nil {
			return err
		}
		if line == "\r\n" {
			// end of header
			break
		}
		line = line[:len(line)-2] // remove \r\n
		result := strings.SplitN(line, ":", 2)
		m.headers[strings.TrimSpace(result[0])] = strings.TrimSpace(result[1])
	}
	return nil
}

func (m *HttpMixin) parseBody(reader *bufio.Reader) error {
	length, ok := m.headers["Content-Length"]
	if ok {
		bodyLen, _ := strconv.Atoi(length)
		m.body = make([]byte, bodyLen)
		reader.Read(m.body)
	}
	return nil
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
	r      *bufio.Reader
	filter *HttpFilter
}

func (d *HttpDecoder) SetReader(r *bufio.Reader) {
	d.r = r
}

func (d *HttpDecoder) Decode() (string, error) {
	if msg, err := d.decodeHttp(); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("%v", msg), nil
	}
}

func (d *HttpDecoder) decodeHttp() (HttpMsg, error) {
	isReq := true
	firstLine, err := d.r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	firstLine = firstLine[:len(firstLine)-2]
	f := strings.SplitN(firstLine, " ", 3)
	fLen := len(f)
	if fLen < 3 {
		return nil, errors.New("bad http msg: " + firstLine)
	} else if fLen == 3 {
		if f[0][:2] == "HT" { // eg: HTTP/1.1 200 OK
			isReq = false
		}
	} else {
		isReq = false
	}
	if isReq {
		// it's http request
		req := new(HttpReq)
		req.method = f[0]
		req.url = f[1]
		req.version = f[2]
		req.parseHeader(d.r)
		req.parseBody(d.r)
		if !d.filter.IsEmpty() && !req.Match(d.filter) {
			return nil, SkipError
		}
		return req, nil
	} else {
		// it's http response
		resp := new(HttpResp)
		resp.version = f[0]
		resp.statusCode, err = strconv.Atoi(f[1])
		if err != nil {
			return nil, errors.New("Invalid http resp: " + firstLine)
		}
		resp.statusMsg = f[2]
		resp.parseHeader(d.r)
		resp.parseBody(d.r)
		if !d.filter.IsEmpty() && !resp.Match(d.filter) {
			return nil, SkipError
		}
		return resp, nil
	}
}

func NewHttpDecoder(filterStr string) *HttpDecoder {
	return &HttpDecoder{
		filter: NewHttpFilter(filterStr)}
}
