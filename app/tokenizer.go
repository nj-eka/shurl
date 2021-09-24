package app

import (
	"errors"
)

var ErrInvalidToken = errors.New("invalid token")

type Tokenizer interface {
	Decode(key string) (int, error)
	Encode(id int) (string, error)
}
