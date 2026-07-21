package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func Read[T any](body io.Reader) (T, error) {
	var decoded T
	err := json.NewDecoder(body).Decode(&decoded)
	if err != nil {
		if _, ok := errors.AsType[*json.SyntaxError](err); ok {
			return *new(T), fmt.Errorf("%w: %w", ErrInvalid, err)
		} else {
			return *new(T), err
		}
	}
	return decoded, nil
}

func Write[T any](w io.Writer, v T) error {
	encoded, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}
	if n, err := w.Write(encoded); err != nil {
		return fmt.Errorf("writing: %w", err)
	} else if encodedLen := len(encoded); n != encodedLen {
		return fmt.Errorf("written %v bytes out of %v", n, encodedLen)
	}
	return nil
}
