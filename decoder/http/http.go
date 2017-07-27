package http

import (
	"github.com/juju/errors"
	"github.com/monsterxx03/pipe/decoder"
	"io"
	"strings"
	"bufio"
	"strconv"
)

var SKIP = errors.New("Skip msg")

type Decoder struct {
	buf    *bufio.Reader
	filter *Filter
}

func (d *Decoder) Decode(reader io.Reader, writer io.Writer) error {
	d.buf = bufio.NewReader(reader)
	for {
		msg, err := d.decodeHttp()
		if err != nil {
			if err == SKIP {
				continue
			}
			return err
		}
		writer.Write([]byte(msg.String()))
	}
	return nil
}

func (d *Decoder) decodeHttp() (Http, error) {
	isReq := true
	firstLine, err := d.buf.ReadString('\n')
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
		req := new(HttpReq)
		req.method = f[0]
		req.url = f[1]
		req.version = f[2]
		req.headers, err = parseHeaders(d.buf)
		if err != nil {
			return nil, err
		}
		req.body = parseBody(req.headers, d.buf)
		if !d.filter.IsEmpty() && !req.Match(d.filter) {
			return nil, SKIP
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
		resp.headers, err = parseHeaders(d.buf)
		if err != nil {
			return nil, err
		}
		resp.body = parseBody(resp.headers, d.buf)
		if !d.filter.IsEmpty() && !resp.Match(d.filter) {
			return nil, SKIP
		}
		return resp, nil
	}
}

func parseHeaders(buf *bufio.Reader) (map[string]string, error) {
	var err error
	var line string
	headers := make(map[string]string)
	for {
		// parse headers
		if line, err = buf.ReadString('\n'); err != nil {
			return nil, err
		}
		if line == "\r\n" {
			// end of header
			break
		}
		line = line[:len(line)-2] // remove \r\n
		result := strings.SplitN(line, ":", 2)
		headers[strings.TrimSpace(result[0])] = strings.TrimSpace(result[1])
	}
	return headers, nil
}

func parseBody(headers map[string]string, reader *bufio.Reader) []byte {
	length, ok := headers["Content-Length"]
	if ok {
		bodyLen, _ := strconv.Atoi(length)
		body := make([]byte, bodyLen)
		reader.Read(body)
		return body
	}
	return nil
}

func (d *Decoder) SetFilter(filter string) {
	d.filter = NewFilter(filter)
}

func init() {
	decoder.Register("http", new(Decoder))
}
