package decoders

import "io"

func DotenvToJson(r io.Reader) (any, error) {
	result, err := EnvToJson(r)
	return result, err
}
