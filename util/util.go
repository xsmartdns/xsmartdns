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
