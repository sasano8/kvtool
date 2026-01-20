package filesystems

import (
	"github.com/sasano8/kvtool/pkg/decoders"
	"github.com/sasano8/kvtool/pkg/sources"
)

type FsFile interface {
	LoadAsJson() (map[string]any, error)
}

type FsEnvFile struct {
}

func (fs *FsEnvFile) LoadAsJson() (map[string]any, error) {
	var err error
	source := sources.SourceEnv{}
	stream, err := source.Load()
	if err != nil {
		return nil, err
	}

	result, err := decoders.EnvToJson(stream)
	if err != nil {
		return nil, err
	}
	return result, nil
}
