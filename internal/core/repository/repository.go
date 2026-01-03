package repository

type Repository[T any] map[string]T

func New[T any]() Repository[T] {
	reg := make(Repository[T])
	return reg
}

func (reg Repository[T]) Get(key string) (T, bool) {
	v, ok := reg[key]
	return v, ok
}

func (reg Repository[T]) Create(key string, value T) (T, bool) {
	if _, ok := reg[key]; ok {
		return reg[key], false
	} else {
		reg[key] = value
		return value, true
	}
}

func (reg Repository[T]) Put(key string, value T) T {
	reg[key] = value
	return value
}

func (reg Repository[T]) Delete(key string) bool {
	if _, ok := reg[key]; ok {
		delete(reg, key)
		return true
	} else {
		return false
	}
}
