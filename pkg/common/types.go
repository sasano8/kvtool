package common

import "io"

type Marshaller interface {
	Marshal(data any) ([]byte, error)
}

// Json互換オブジェクトから Encoder を通して別の表現として出力するための struct
type marshalizable struct {
	Encoder   Marshaller
	Src       any
	marshaled []byte
}

func NewMarshaller(encoder Marshaller, src any) marshalizable {
	return marshalizable{
		Encoder: encoder,
		Src:     src,
	}
}

func (data *marshalizable) Marshal() ([]byte, error) {
	if data.marshaled == nil {
		b, err := data.Encoder.Marshal(data.Src)
		if err != nil {
			return nil, err
		}
		data.marshaled = b
		return data.marshaled, nil
	} else {
		return data.marshaled, nil
	}
}

func (data *marshalizable) Dump(w io.Writer) error {
	b, err := data.Marshal()
	if err != nil {
		return err
	} else {
		_, err = w.Write(b)
		return err
	}
}
