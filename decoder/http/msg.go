package http

import (
	"bytes"
	"fmt"
	"github.com/ugorji/go/codec"
	"reflect"
	"regexp"
	"strconv"
)

var mh codec.MsgpackHandle

type Http interface {
	Match(*Filter) bool
	StringHeader() string
	DecodeBody() (string, error)
	RawBody() []byte
}

func prettyPrint(v map[string]interface{}) {
	for key, value := range v {
		keyType := reflect.ValueOf(key).Kind()
		valueType := reflect.ValueOf(value).Kind()
	}
}

func decodeToString(contentType string, data []byte) (string, error) {
	var v map[string]interface{}
	switch contentType {
	case "application/msgpack":
		dec := codec.NewDecoder(bytes.NewReader(data), &mh)
		err := dec.Decode(&v)
		if err != nil {
			return "", err
		}
		prettyPrint(v)
		return fmt.Sprint(v), nil
	}
	return string(data), nil
}

type HttpReq struct {
	method  string
	url     string
	version string
	headers map[string]string
	body    []byte
}

func (m *HttpReq) RawBody() []byte {
	return m.body
}

func (m *HttpReq) DecodeBody() (string, error) {
	return decodeToString(m.headers["content-type"], m.body)
}

func (m *HttpReq) StringHeader() string {
	headStr := ""
	for k, v := range m.headers {
		headStr += k + ": " + v + "\r\n"
	}
	return fmt.Sprintf("%s %s %s\r\n%s\r\n", m.method, m.url, m.version, headStr)
}

func (m *HttpReq) Match(filter *Filter) bool {
	filters := make(map[string]*regexp.Regexp)
	mapCopy(filters, filter.filters)
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
	headers    map[string]string
	body       []byte
}

func (m *HttpResp) RawBody() []byte {
	return m.body
}

func (m *HttpResp) DecodeBody() (string, error) {
	return decodeToString(m.headers["content-type"], m.body)
}

func (m *HttpResp) StringHeader() string {
	headStr := ""
	for k, v := range m.headers {
		headStr += k + ": " + v + "\r\n"
	}
	return fmt.Sprintf("%s %s %s\r\n%s\r\n", m.version, m.statusCode, m.statusMsg, headStr)
}

func (m *HttpResp) Match(filter *Filter) bool {
	filters := make(map[string]*regexp.Regexp)
	mapCopy(filters, filter.filters)
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

func mapCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
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
