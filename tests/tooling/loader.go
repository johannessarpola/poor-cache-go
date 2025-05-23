package tooling

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

type Source interface {
	Next() (string, bool)
}

var _ Source = (*KeySource)(nil)

type KeySource struct {
	index int
	keys  []string
	mu    *sync.Mutex
}

// Next implements Source.
func (k *KeySource) Next() (string, bool) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.index >= len(k.keys) {
		k.index = 0 // reset index
	}

	key := k.keys[k.index]
	k.index++
	return key, true
}

func New(keys []string) *KeySource {
	return &KeySource{
		index: 0,
		keys:  keys,
		mu:    &sync.Mutex{},
	}
}

func UnmarshalFrom(r io.Reader) (*KeySource, error) {
	var keys []string
	err := json.NewDecoder(r).Decode(&keys)
	if err != nil {
		return nil, err
	}
	return New(keys), nil
}

func LoadFrom(file string) (*KeySource, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return UnmarshalFrom(f)
}
