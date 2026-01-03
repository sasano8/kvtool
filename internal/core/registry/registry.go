package registry

type Registry[T any] map[string]T

func New[T any]() Registry[T] {
	reg := make(Registry[T])
	return reg
}

func (reg Registry[T]) Get(key string) (T, bool) {
	v, ok := reg[key]
	return v, ok
}

func (reg Registry[T]) Create(key string, value T) (T, bool) {
	if _, ok := reg[key]; ok {
		return reg[key], false
	} else {
		reg[key] = value
		return value, true
	}
}

func (reg Registry[T]) Put(key string, value T) T {
	reg[key] = value
	return value
}

func (reg Registry[T]) Delete(key string) bool {
	if _, ok := reg[key]; ok {
		delete(reg, key)
		return true
	} else {
		return false
	}
}
