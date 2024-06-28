package util

import "encoding/json"

func Copy[T any](v T) (T, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return v, err
	}
	tmp := new(T)
	if err = json.Unmarshal(b, tmp); err != nil {
		return v, err
	}
	return *tmp, nil
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
