package repository

import (
	"errors"
	"fmt"
	"strings"
)

type Repository[T any] map[string]T

func New[T any]() Repository[T] {
	reg := make(Repository[T])
	return reg
}

var (
	ErrNotFound  = errors.New("key is not found")
	ErrKeyExists = errors.New("key is already exists")
	ErrEmptyPath = errors.New("key is empty")
	ErrBadChar   = errors.New("key contains invalid character")
	// ErrNullByte    = errors.New("key contains NUL byte")
	// ErrControlChar = errors.New("key contains control character")
)

func ValidatePath(p string) error {
	const invalid = "<>\"|?*/"
	/*
		/: Repositoryをネストさせる際に / で区切るので禁止
	*/
	if p == "" {
		return fmt.Errorf("%w", ErrEmptyPath)
	}

	for _, r := range p {
		if strings.ContainsRune(invalid, r) {
			return fmt.Errorf("%w: %q contains %q", ErrBadChar, p, r)
		}
	}
	return nil
}

func (reg Repository[T]) Get(key string) (T, error) {
	if v, ok := reg[key]; ok {
		return v, nil
	} else {
		var zero T
		return zero, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
}

func (reg Repository[T]) Create(key string, value T) (T, error) {
	if _, ok := reg[key]; ok {
		var zero T
		return zero, fmt.Errorf("%w: %s", ErrKeyExists, key)
	} else {
		v, err := reg.Put(key, value)
		return v, err
	}
}

func (reg Repository[T]) Put(key string, value T) (T, error) {
	err := ValidatePath(key)
	if err == nil {
		reg[key] = value
		return value, nil
	} else {
		var zero T
		return zero, err
	}
}

func (reg Repository[T]) Delete(key string) error {
	if _, ok := reg[key]; ok {
		delete(reg, key)
		return nil
	} else {
		return fmt.Errorf("%w: %s", ErrNotFound, key)
	}
}
