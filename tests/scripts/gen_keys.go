package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/johannessarpola/poor-cache-go/tests/generators"
)

const (
	seedn int64 = 42
	len   int8  = 12
)

func main() {
	// Define a command-line flag for the number of keys
	numKeys := flag.Int("num", 100, "Number of keys to generate")
	flag.IntVar(numKeys, "n", 100, "Number of keys to generate (shorthand)")
	outputFile := flag.String("out", "", "Output file to save the keys")
	flag.StringVar(outputFile, "o", "", "Output file to save the keys (shorthand)")
	flag.Parse()

	seed := rand.NewSource(seedn)
	// Generate the specified number of keys
	keys := make([]string, *numKeys)
	for i := 0; i < *numKeys; i++ {
		key := generators.RandString(len, seed)
		keys[i] = key
	}

	// Convert the keys to JSON
	keysJSON, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error converting keys to JSON:", err)
		os.Exit(1)
	}

	// Output the JSON to a file or standard output
	if *outputFile != "" {
		err = os.WriteFile(*outputFile, keysJSON, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing to file:", err)
			os.Exit(1)
		}
		fmt.Println("Keys written to", *outputFile)
	} else {
		fmt.Println(string(keysJSON))
	}
}
