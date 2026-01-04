package repository

import (
	"context"
	"fmt"
	"time"
)

type FS[T any] map[string]T

func ResolveFs[T any](fs FS[T], path string) (FS[T], string, error) {
	parentContext := context.Background()
	timeoutContext, cancel := context.WithTimeout(parentContext, 5*time.Second)
	defer cancel()

	resolved_fs := fs
	remain_path := path
	var err error

	for {
		if err == ErrSuccess {
			break
		}

		select {
		case <-timeoutContext.Done():
			return nil, "", fmt.Errorf("Operation timed out")
		default:
			resolved_fs, remain_path, err = resolvedFs(parentContext, resolved_fs, remain_path)
		}
	}

	if err == ErrSuccess {
		return resolved_fs, remain_path, nil
	} else {
		if err == nil {
			panic("Invalid status.")
		}
		return nil, remain_path, err
	}
}

func resolvedFs[T any](ctx context.Context, fs FS[T], path string) (FS[T], string, error) {
	resolved_fs := fs
	return resolved_fs, "remain_path", fmt.Errorf("")
}
