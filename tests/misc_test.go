package tests

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"testing"
)

// Your function to benchmark
func BenchmarkPrinter(b *testing.B) {
	i := 0
	for b.Loop() {
		fmt.Printf("Hello, World - %d\n", i)
		i++
	}
}

func BenchmarkParallePrinter(b *testing.B) {
	templ := template.Must(template.New("test").Parse("Hello, {{.Name}} {{.Iteration}}!"))
	// Create a template context with the named argument
	type data struct {
		Name      string
		Iteration int
	}

	i := 0
	mu := sync.Mutex{}
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			buf.Reset()
			n := i
			mu.Lock()
			i++
			mu.Unlock()
			d := &data{Iteration: n, Name: "World"}
			templ.Execute(&buf, d)
		}
	})
}
