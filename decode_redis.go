package main

import (
	"bufio"
	"fmt"
	_ "log"
	"strconv"
	"strings"
)

const (
	respOK     = '+'
	respERROR  = '-'
	respInt    = ':'
	respString = '$'
	respArray  = '*'
)

type RedisDecoder struct {
	r *bufio.Reader
}

func (d *RedisDecoder) SetReader(r *bufio.Reader) {
	d.r = r
}

type RedisMsg interface{}

func NewRedisDecoder(filterStr string) *RedisDecoder {
	return new(RedisDecoder)
}

func (d *RedisDecoder) Decode() (string, error) {
	if result, err := d.decodeRedisMsg(); err != nil {
		return "", err
	} else {
		switch v := result.(type) {
		case nil:
			return "nil", nil
		case int:
			return strconv.Itoa(v), nil
		case string:
			return v, nil
		case []RedisMsg:
			_result := make([]string, len(v))
			for i, item := range v {
				_result[i] = fmt.Sprint(item)
			}
			return strings.Join(_result, " "), nil
		default:
			return fmt.Sprint(v), nil
		}
	}
}

func (d *RedisDecoder) decodeRedisMsg() (RedisMsg, error) {
	line, err := d.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = line[:len(line)-2] // truncate end \r\n
	headerByte, resp := line[0], line[1:]
	switch headerByte {
	case respOK:
		return resp, nil
	case respERROR:
		return resp, nil
	case respInt:
		if intValue, err := strconv.Atoi(resp); err != nil {
			return nil, err
		} else {
			return intValue, nil
		}
	case respString:
		strLen, err := strconv.Atoi(resp)
		if err != nil {
			return "", err
		}
		if strLen == -1 {
			return nil, nil
		}
		line, _ = d.r.ReadString('\n')
		return line[:len(line)-2], nil
	case respArray:
		arrayLen, err := strconv.Atoi(resp)
		if err != nil {
			return "", err
		}
		if arrayLen == -1 {
			// empty array
			return nil, nil
		}
		result := make([]RedisMsg, arrayLen)
		for i := 0; i < arrayLen; i++ {
			if result[i], err = d.Decode(); err != nil {
				return "", err
			}
		}
		return result, nil
	default:
		return nil, nil
	}
}
