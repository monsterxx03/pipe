package redis

import (
	"io"
	"github.com/monsterxx03/pipe/decoder"
)

type Decoder struct {

}

func (d *Decoder) Decode(reader io.Reader, writer io.Writer) {

}

func (d *Decoder) SetFilter(filter string) {

}

func init() {
	decoder.Register("redis", new(Decoder))
}
