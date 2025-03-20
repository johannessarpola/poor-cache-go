package common

import "time"

type Meta struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type Value struct {
	Meta Meta
	Data []byte
}

type Item struct {
	Value      Value
	Expiration time.Time
}
