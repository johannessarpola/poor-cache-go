package rest

import (
	"time"

	"github.com/johannessarpola/poor-cache-go/internal/common"
)

type Store interface {
	Set(key string, value any, ttl time.Duration) error
	Get(key string, dest any) (*common.Meta, error)
	Delete(key string) error
	Has(key string) bool
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{
		store,
	}
}
