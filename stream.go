package main

import (
	"io"
	"github.com/monsterxx03/pipe/decoder"
)

type Stream struct {
	pr      *io.PipeReader
	pw      *io.PipeWriter
	decoder decoder.Decoder
}

func (s *Stream) Write(data []byte) (int, error) {
	return s.pw.Write(data)
}

func (s *Stream) Read(data []byte) (int, error) {
	return s.pr.Read(data)
}


func (s *Stream) To(w io.Writer) {
	s.decoder.Decode(s.pr, w)
}

func NewStream(decoder decoder.Decoder) *Stream {
	pr, pw :=  io.Pipe()
	s := &Stream{pr, pw, decoder}
	return s
}
