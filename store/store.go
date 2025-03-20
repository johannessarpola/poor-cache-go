package store

import (
	"fmt"
	"sync"
	"time"
)

type Meta struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type Value struct {
	Meta Meta
	Data []byte
}
type item struct {
	value      Value
	expiration time.Time
}

type Store struct {
	mu              sync.RWMutex
	data            map[string]item
	cleanupInterval time.Duration
	cleanupQueue    chan string
	mainQuit        chan struct{}
	subQuits        []chan struct{} // this is to broadcast the quit signal to all subroutines
	wg              *sync.WaitGroup
}

type Option func(*Store)

func WithCleanupInterval(interval time.Duration) Option {
	return func(s *Store) {
		s.cleanupInterval = interval
	}
}

const bufferSize = 32

func New(opt ...Option) *Store {
	wg := &sync.WaitGroup{}
	s := &Store{
		wg:              wg,
		data:            make(map[string]item),
		cleanupInterval: 1 * time.Minute,
		cleanupQueue:    make(chan string, bufferSize),
		mainQuit:        make(chan struct{}, 1),
	}

	for _, o := range opt {
		o(s)
	}

	q1 := make(chan struct{}, 1)
	s.subQuits = append(s.subQuits, q1)
	go s.cleanupExpiredKeys(q1)
	wg.Add(1)

	q2 := make(chan struct{}, 1)
	s.subQuits = append(s.subQuits, q2)
	go s.takekOutTheTrash(q2)
	wg.Add(1)

	go s.broadcastQuits()
	wg.Add(1)
	return s
}

func (s *Store) broadcastQuits() {
	<-s.mainQuit
	for _, q := range s.subQuits {
		q <- struct{}{}
		close(q)
	}
	s.wg.Done()
}

func (s *Store) Set(key string, value any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, err := Serialize(value)
	if err != nil {
		return err
	}

	meta := Meta{
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	expiration := time.Now().Add(ttl)
	s.data[key] = item{value: Value{Meta: meta, Data: v}, expiration: expiration}
	return nil
}

func (s *Store) Get(key string, dest any) (*Meta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.data[key]
	if !exists {
		return nil, nil
	}

	if time.Now().After(item.expiration) {
		s.cleanupQueue <- key // This should not block since the channel is buffered.
		return nil, nil
	}

	meta := item.value.Meta
	err := Deserialize(item.value.Data, dest)
	return &meta, err
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *Store) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists
}

func (s *Store) takekOutTheTrash(quit chan struct{}) {
	for {
		select {
		case <-quit:
			fmt.Println("Trashman exiting...")
			s.wg.Done()
			return
		case v, ok := <-s.cleanupQueue:
			if !ok {
				return
			}
			s.Delete(v)
		}
	}
}

func (s *Store) cleanupExpiredKeys(quit chan struct{}) {
	ticker := time.Tick(s.cleanupInterval)
	timeouter := make(chan struct{}, 1)
	for {
		select {
		case <-quit:
			timeouter <- struct{}{}
			fmt.Println("Observer exiting...")
			s.wg.Done()
			return
		case <-ticker:
			for key, data := range s.data {
				if time.Now().After(data.expiration) {
					select {
					case <-timeouter:
					case s.cleanupQueue <- key:
					}
				}
			}
		}
	}
}

func (s *Store) Close() {
	s.mainQuit <- struct{}{}
	close(s.mainQuit)
	s.wg.Wait()
}
