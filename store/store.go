package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/johannessarpola/poor-cache-go/pb"
	"google.golang.org/protobuf/proto"
	proto_time "google.golang.org/protobuf/types/known/timestamppb"
)

type Meta struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func ExtractMeta(value *pb.Value) Meta {

	meta := Meta{
		CreatedAt:  value.CreatedAt.AsTime(),
		ModifiedAt: value.ModifiedAt.AsTime(),
	}
	return meta
}

type item struct {
	value      []byte
	expiration time.Time
}

type Store struct {
	mu              sync.RWMutex
	data            map[string]item
	cleanupInterval time.Duration
	cleanupQueue    chan string
	mainQuit        chan struct{}
	subQuits        []chan struct{}
}

type Option func(*Store)

func WithCleanupInterval(interval time.Duration) Option {
	return func(s *Store) {
		s.cleanupInterval = interval
	}
}

const bufferSize = 32

func New(opt ...Option) *Store {
	s := &Store{
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
	q2 := make(chan struct{}, 1)
	s.subQuits = append(s.subQuits, q2)
	go s.takekOutTheTrash(q2)
	go s.broadcastQuits()
	return s
}

func (s *Store) broadcastQuits() {
	<-s.mainQuit
	for _, q := range s.subQuits {
		q <- struct{}{}
		close(q)
	}
}

func (s *Store) Set(key string, value any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, err := Serialize(value)
	if err != nil {
		return err
	}

	now := proto_time.Now()
	if v.CreatedAt == nil {
		v.CreatedAt = now // Set creation time if not set
	}
	v.ModifiedAt = now // Update modified time

	serialized, err := proto.Marshal(v)
	if err != nil {
		return err
	}

	expiration := time.Now().Add(ttl)
	s.data[key] = item{value: serialized, expiration: expiration}
	return nil
}

func (s *Store) Get(key string, dest any) (*Meta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.data[key]
	if !exists {
		return nil, nil
	}

	if time.Now().After(data.expiration) {
		return nil, nil
	}

	var pbValue pb.Value
	if err := proto.Unmarshal(data.value, &pbValue); err != nil {
		return nil, err
	}

	meta := ExtractMeta(&pbValue)
	err := Deserialize(&pbValue, dest)
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
}
