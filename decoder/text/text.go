package text

import (
	"github.com/monsterxx03/pipe/decoder"
	"io"
)

type Decoder struct {

}

func (d *Decoder) Decode(reader io.Reader, writer io.Writer) {
	io.Copy(writer, reader)
}

func (d *Decoder) SetFilter(filter string) {

}


func init() {
	decoder.Register("text", new(Decoder))
}