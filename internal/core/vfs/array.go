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
		case <-parentContext.Done():
			return nil, "", fmt.Errorf("Operation timed out")
		case <-timeoutContext.Done():
			return nil, "", fmt.Errorf("Operation timed out")
		default:
			if true {
				resolved_fs, remain_path, err = fs.Resolve(parentContext, remain_path)
			} else {
				// Resolve が実装されていない場合
				resolved_fs = nil
				err = repository.ErrSuccess
			}
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
