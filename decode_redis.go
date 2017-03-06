package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	_ "log"
	"strconv"
	_ "strings"
)

const (
	respOK     = '+'
	respERROR  = '-'
	respInt    = ':'
	respString = '$'
	respArray  = '*'
)

func decodeRedis(data []byte) (interface{}, error) {
	buf := bytes.NewBuffer(data)

	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF && len(line) == 0 {
			break
		}
		line = line[:len(line)-2] // truncate end \r\n
		headerByte, resp := line[0], line[1:]
		switch headerByte {
		case respOK:
			return resp, nil
		case respERROR:
			return errors.New(resp), nil
		case respInt:
			return strconv.ParseInt(resp, 10, 64)
		case respString:
			strLen, err := strconv.Atoi(resp)
			if err != nil {
				return nil, err
			}
			if strLen == -1 {
				return nil, nil
			}
			line, _ = buf.ReadString('\n')
			return string(line[:len(line)-2]), nil
		case respArray:
			arrayLen, _ := strconv.Atoi(resp)
			result := make([]string, arrayLen)
			for i := 0; i < arrayLen; i++ {
				line, _ := buf.ReadString('\n')
				subResult, err := decodeRedis([]byte(line[:len(line)-2]))
				if err != nil {
					return nil, err
				}
				// TODO convert subResult to []byte
				result = append(result, subResult)
			}
			return result, nil
		default:
			fmt.Println("unknow")
		}

		if err == io.EOF {
			break
		}
	}
	return nil, nil
}
