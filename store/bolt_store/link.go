package bolt_store

import "time"

type Link struct {
	Id        int    `storm:"id,increment"`
	TargetUrl string `storm:"unique"`
	CreatedAt time.Time
	DeletedAt *time.Time
	ExpiredAt *time.Time
	Hits      int
}
