package http

import (
	"github.com/monsterxx03/pipe/decoder"
	"io"
)

type Decoder struct {

}

func (d *Decoder) Decode(reader io.Reader, writer io.Writer) {

}

func (d *Decoder) SetFilter(filter string) {

}

func init() {
	decoder.Register("http", new(Decoder))
}
