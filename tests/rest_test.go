package tests

import (
	"bytes"
	"math/rand"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "embed"

	"github.com/johannessarpola/poor-cache-go/tests/tooling"
)

//go:embed data.json
var data []byte

const (
	host    = "http://localhost:8080"
	apiBase = "api/v1"
)

// Mock data for operations
var mockData = []byte(`{"value": "test_value"}`)

func combine(strs ...string) string {
	return strings.Join(strs, "/")
}

func newRequest(op int, key string, body []byte) *http.Request {
	setPath := combine(host, apiBase, "set")
	getPath := combine(host, apiBase, "get")
	deletePath := combine(host, apiBase, "delete")
	hasPath := combine(host, apiBase, "has")

	switch op {
	case 0:
		r, _ := http.NewRequest(
			"POST",
			combine(setPath, key),
			bytes.NewBuffer(body))
		return r
	case 1:
		r, _ := http.NewRequest(
			"GET",
			combine(getPath, key),
			nil)
		return r
	case 2:
		r, _ := http.NewRequest(
			"DELETE",
			combine(deletePath, key),
			nil)
		return r
	case 3:
		r, _ := http.NewRequest(
			"GET",
			combine(hasPath, key),
			nil)
		return r
	}
	return nil
}

func BenchmarkRestAPIShotgun(b *testing.B) {
	const mrn = 4
	const seedn = 42
	seed := rand.NewSource(seedn)
	r := rand.New(seed)

	bb := bytes.NewBuffer(data)
	source, err := tooling.UnmarshalFrom(bb)
	if err != nil {
		b.Errorf("Failed to load source: %v", err)
	}

	var errs atomic.Int64
	var ok atomic.Int64
	var failure atomic.Int64
	var latency atomic.Int64

	client := &http.Client{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			op := r.Intn(mrn)
			key, _ := source.Next()
			var req *http.Request
			if op == 0 {
				req = newRequest(op, key, mockData) // TODO Cleanup
			} else {
				req = newRequest(op, key, nil)
			}

			start := time.Now()
			resp, err := client.Do(req)
			duration := time.Since(start).Milliseconds()

			if err != nil {
				errs.Add(1)
			}
			latency.Add(duration)
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				ok.Add(1)
			} else {
				failure.Add(1)
			}

			resp.Body.Close()

		}
	}) // end runParallel'

	b.ReportMetric(float64(latency.Load())/float64(b.N), "latency_ms/op") // Report average latency per request
	b.ReportMetric(float64(errs.Load())/float64(b.N), "errors/op")
	b.ReportMetric(float64(ok.Load())/float64(b.N), "ok/op")
	b.ReportMetric(float64(failure.Load())/float64(b.N), "failure/op")
}
