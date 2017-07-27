package redis

import (
	"io"
	"github.com/monsterxx03/pipe/decoder"
	"bufio"
	"github.com/juju/errors"
)

const (
	respOK     = '+'
	respERROR  = '-'
	respInt    = ':'
	respString = '$'
	respArray  = '*'
)
var NIL = []byte("nil")

type RedisMsg interface{}

type Decoder struct {
	buf *bufio.Reader
}

func (d *Decoder) Decode(reader io.Reader, writer io.Writer) error {
	d.buf = bufio.NewReader(reader)
	for {
		if result, err := d.decodeRedisMsg(); err != nil {
			return err
		} else {
			writer.Write(result)
			writer.Write([]byte("\n"))
		}
	}
	return nil
}

func (d *Decoder) SetFilter(filter string) {

}


func (d *Decoder) decodeRedisMsg() ([]byte, error) {
	line, err := d.buf.ReadBytes('\n')
	if err != nil {
		return []byte{}, err
	}
	line = line[:len(line)-2] // truncate end \r\n
	headerByte, resp := line[0], line[1:]
	switch headerByte {
	case respOK:
	case respERROR:
	case respInt:
		return resp, nil
	case respString:
		strLen, err := parseLen(resp)
		if err != nil {
			return nil, err
		}
		if strLen == -1 {
			return NIL, nil
		}
		line, _ = d.buf.ReadBytes('\n')
		return line[:len(line)-2], nil
	case respArray:
		arrayLen, err := parseLen(resp)
		if err != nil {
			return nil, err
		}
		if arrayLen == -1 {
			// empty array
			return NIL, nil
		}
		result := []byte{}
		for i := 0; i < arrayLen; i++ {
			tmp, _ := d.decodeRedisMsg()
			result = append(result, tmp...)
			result = append(result, byte(' '))
		}
		return result, nil
	default:
		return nil, nil
	}
	return nil, nil
}

func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, errors.New("malformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, errors.New("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}


func init() {
	decoder.Register("redis", new(Decoder))
}
