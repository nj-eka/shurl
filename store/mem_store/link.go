package mem_store

import (
	"time"
)

type Link struct {
	Id        int        `json:"id"`
	TargetUrl string     `json:"url"`
	CreatedAt time.Time  `json:"ct"`
	DeletedAt *time.Time `json:"dt"`
	ExpiredAt *time.Time `json:"et"`
	Hits      int        `json:"hs"`
}
