package udp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/johannessarpola/poor-cache-go/internal/common"
	"github.com/johannessarpola/poor-cache-go/internal/logger"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' {
		sd := string(b[1 : len(b)-1])
		d.Duration, err = time.ParseDuration(sd)
		return
	}

	var id int64
	id, err = json.Number(string(b)).Int64()
	d.Duration = time.Duration(id)

	return
}

func (d Duration) MarshalJSON() (b []byte, err error) {
	if d.Duration == 0 {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

type Envelope struct {
	Cmd     string   `json:"cmd"`
	Key     string   `json:"key,omitempty"`
	Value   any      `json:"value,omitempty"`
	Error   string   `json:"error,omitempty"`
	Success bool     `json:"succes"`
	TTL     Duration `json:"ttl,omitzero"`
}

type Server struct {
	addr           *net.UDPAddr
	conn           *net.Conn
	mainQuit       chan struct{}
	handlerTimeout time.Duration
	store          Store
	wg             *sync.WaitGroup
}

type Store interface {
	Set(key string, value any, ttl time.Duration) error
	Get(key string, dest any) (*common.Meta, error)
	Delete(key string) error
	Has(key string) bool
}

func New(address string, port int, store Store) *Server {

	return &Server{
		addr: &net.UDPAddr{
			Port: port,
			IP:   net.ParseIP(address),
		},
		conn:           nil,
		mainQuit:       make(chan struct{}, 1),
		handlerTimeout: 5 * time.Second,
		store:          store,
		wg:             &sync.WaitGroup{},
	}
}

// TODO Use context from parent
func (s *Server) Start() error {

	conn, err := net.ListenUDP("udp", s.addr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
		return err
	}
	logger.Infof("Started UDP listener at %s", s.addr.String())
	defer conn.Close()
	buffer := make([]byte, 2048)
	for {
		select {
		case <-s.mainQuit:
			return nil
		default:
			ctx, cancel := context.WithTimeout(context.Background(), s.handlerTimeout)
			defer cancel()
			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading from UDP:", err)
				continue
			}
			s.wg.Add(1)
			go handleRequest(ctx, conn, clientAddr, buffer[:n], s.store, s.wg.Done)

		}

	}

}

func handleRequest(ctx context.Context, conn *net.UDPConn, clientAddr *net.UDPAddr, buffer []byte, store Store, onDone func()) {
	defer onDone()
	var envelope Envelope
	err := json.Unmarshal(buffer, &envelope)
	if err != nil {
		logger.Errorf("Error unmarshalling request: %s", err)
		return
	}
	// TODO Handle errors to response
	switch envelope.Cmd {
	case "SET":
		store.Set(envelope.Key, envelope.Value, envelope.TTL.Duration)
		response := Envelope{
			Cmd:     envelope.Cmd,
			Success: true,
		}
		res, err := json.Marshal(response)
		if err != nil {
			logger.Errorf("Error marshalling response: %s", err)
			return
		}
		_, err = conn.WriteToUDP(res, clientAddr)
		if err != nil {
			logger.Errorf("Error writing to UDP: %s", err)
			return
		}

	case "GET":
		var dest any
		meta, err := store.Get(envelope.Key, &dest)
		if err != nil {
			logger.Errorf("Error getting key: %s", err)
			return
		}
		data := make(map[string]any)
		data["meta"] = meta
		data["data"] = dest

		response := Envelope{
			Cmd:     envelope.Cmd,
			Success: meta != nil,
			Value:   data,
		}
		res, err := json.Marshal(response)
		if err != nil {
			logger.Errorf("Error marshalling response: %s", err)
			return
		}
		_, err = conn.WriteToUDP(res, clientAddr)
		if err != nil {
			logger.Errorf("Error writing to UDP: %s", err)
			return
		}
	case "DELETE":
		store.Delete(envelope.Key)
		response := Envelope{
			Cmd:     envelope.Cmd,
			Success: true,
		}
		res, err := json.Marshal(response)
		if err != nil {
			logger.Errorf("Error marshalling response: %s", err)
			return
		}
		_, err = conn.WriteToUDP(res, clientAddr)
		if err != nil {
			logger.Errorf("Error writing to UDP: %s", err)
			return
		}
	case "HAS":
		has := store.Has(envelope.Key)
		response := Envelope{
			Cmd:   envelope.Cmd,
			Value: has,
		}
		res, err := json.Marshal(response)
		if err != nil {
			logger.Errorf("Error marshalling response: %s", err)
			return
		}
		_, err = conn.WriteToUDP(res, clientAddr)
		if err != nil {
			logger.Errorf("Error writing to UDP: %s", err)
			return
		}
	default:
		logger.Errorf("Unknown command: %s", envelope.Cmd)
	}

}

func (s *Server) Close() {
	s.mainQuit <- struct{}{}
	s.wg.Wait() // TODO This should timeout
}
