package main

import (
	"github.com/monsterxx03/pipe/decoder"
	"io"
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
	opts := new(decoder.Options)
	if *deepDecode != "" {
		opts.DeepDecode = true
	}
	if err := s.decoder.Decode(s.pr, w, opts); err != nil {
		panic(err)
	}
}

func NewStream(decoder decoder.Decoder) *Stream {
	pr, pw := io.Pipe()
	s := &Stream{pr, pw, decoder}
	return s
}
