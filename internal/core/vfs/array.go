package repository

import (
	"context"
	"fmt"
	"time"

	repository "github.com/sasano8/kvtool/internal/core/repositories"
)

func ResolveFs(fs repository.Repository[any], path string) (repository.Repository[any], string, error) {
	parentContext := context.Background()
	timeoutContext, cancel := context.WithTimeout(parentContext, 5*time.Second)
	defer cancel()

	resolved_fs := fs
	remain_path := path
	var err error

	for {
		if err == repository.ErrSuccess {
			break
		}

		select {
		case <-timeoutContext.Done():
			return nil, "", fmt.Errorf("Operation timed out")
		default:
			resolved_fs, remain_path, err = fs.Resolve(parentContext, remain_path)
		}
	}

	if err == repository.ErrSuccess {
		return resolved_fs, remain_path, nil
	} else {
		if err == nil {
			panic("Invalid status.")
		}
		return nil, remain_path, err
	}
}
