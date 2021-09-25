package hashid_tokenizer

import (
	"fmt"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/config"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/logging"
	"github.com/speps/go-hashids"
)

type HashidTokenizer struct {
	h *hashids.HashID
}

func (htz *HashidTokenizer) Decode(key string) (int, error) {
	if ids, err := htz.h.DecodeWithError(key); err != nil {
		return -1, err
	} else {
		if len(ids) != 1 {
			return -1, fmt.Errorf("unexpected decoding results %s -> %v: %w", key, ids, app.ErrInvalidToken)
		}
		return ids[0], nil
	}
}

func (htz *HashidTokenizer) Encode(id int) (string, error) {
	if key, err := htz.h.Encode([]int{id}); err != nil {
		return "", err
	} else {
		return key, nil
	}
}

func NewHashidTokenizer(cfg *config.HashidTokenizerConfig) (app.Tokenizer, error) {
	logging.Msg(cu.Operation("hashid_init")).Debugf("config: %v", cfg)
	hd := hashids.NewData()
	if cfg != nil {
		if cfg.Salt != "" {
			hd.Salt = cfg.Salt
		}
		if cfg.MinLength > 0 {
			hd.MinLength = cfg.MinLength
		}
		if cfg.Alphabet != "" {
			hd.Alphabet = cfg.Alphabet
		}
	}
	if h, err := hashids.NewWithData(hd); err == nil {
		return &HashidTokenizer{h: h}, nil
	} else {
		return nil, err
	}
}
