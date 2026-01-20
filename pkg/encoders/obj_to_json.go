package encoders

import (
	"bytes"
	"encoding/json"
)

type ObjToJsonEncoder struct {
	Newline bool
}

func NewJsonEncoder() ObjToJsonEncoder {
	return ObjToJsonEncoder{
		Newline: false,
	}
}

func (encoder *ObjToJsonEncoder) Marshal(data any) ([]byte, error) {
	if encoder.Newline {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(data); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	} else {
		// Marshal は細かい制御ができない
		b, err := json.Marshal(data)
		return b, err
	}
}
