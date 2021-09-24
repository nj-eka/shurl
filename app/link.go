package app

import (
	"errors"
	"time"
)

var ErrInvalidUrl = errors.New("invalid url")

type Link struct {
	Id        int
	Key       string
	TargetUrl string
	CreatedAt time.Time
	ExpiredAt *time.Time
	DeletedAt *time.Time
	Hits      int
}
