package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	const (
		webDir = "web"
		addr   = ":8080"
	)

	if _, err := os.Stat(webDir); err != nil {
		log.Fatalf("missing web assets: %v", err)
	}

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	log.Printf("listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
