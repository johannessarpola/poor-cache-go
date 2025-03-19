package main

import (
	"fmt"
	"log"
	"time"

	"github.com/johannessarpola/poor-cache-go/store"
)

func main() {
	store := store.New()

	// Define a struct
	type User struct {
		Name  string
		Age   int
		Email string
	}

	user := User{Name: "Alice", Age: 30, Email: "alice@example.com"}

	if err := store.Set("user:alice", user, 2*time.Second); err != nil {
		log.Fatal(err)
	}

	var retrievedUser User
	if _, err := store.Get("user:alice", &retrievedUser); err != nil {
		log.Fatal(err)
	}

	type Document struct {
		Headline string `json:"headline"`
		Text     string `json:"text"`
	}

	doc := Document{Headline: "Hello, World!", Text: "This is a sample document."}
	// Store the struct
	if err := store.Set("doc:hello", doc, 2*time.Second); err != nil {
		log.Fatal(err)
	}
	var retrievedDoc Document
	if _, err := store.Get("doc:hello", &retrievedDoc); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved User: %+v\n", retrievedUser)
	fmt.Printf("Retrieved Doc: %+v\n", retrievedDoc)

	// test expiry
	var expiredDoc Document
	time.Sleep(5 * time.Second)
	if _, err := store.Get("doc:hello", &expiredDoc); err != nil {
		log.Fatal(err)
	}
	fmt.Println(expiredDoc) // should be empty

	store.Close() // This is not so simple but lets just do this for now.
	time.Sleep(1 * time.Second)
}
