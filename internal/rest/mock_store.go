package rest

import (
	"time"

	"github.com/johannessarpola/poor-cache-go/internal/common"
)

type MockStore struct {
	SetFunc    func(key string, value any, ttl time.Duration) error
	GetFunc    func(key string, dest any) (*common.Meta, error)
	DeleteFunc func(key string) error
	HasFunc    func(key string) bool
}

func (m *MockStore) Set(key string, value any, ttl time.Duration) error {
	return m.SetFunc(key, value, ttl)
}

func (m *MockStore) Get(key string, dest any) (*common.Meta, error) {
	return m.GetFunc(key, dest)
}

func (m *MockStore) Delete(key string) error {
	return m.DeleteFunc(key)
}

func (m *MockStore) Has(key string) bool {
	return m.HasFunc(key)
}
