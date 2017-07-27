package text

import (
	"github.com/monsterxx03/pipe/decoder"
	"io"
)

type Decoder struct {

}

func (d *Decoder) Decode(reader io.Reader, writer io.Writer) error {
	_, err := io.Copy(writer, reader)
	return err
}

func (d *Decoder) SetFilter(filter string) {

}


func init() {
	decoder.Register("text", new(Decoder))
}