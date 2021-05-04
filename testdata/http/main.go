package main

import (
	"log"
	"net/http"
	"os"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	// Register handler function
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent) // no content
	})

	// Register static assets handler
	http.Handle("/assets", http.FileServer(http.Dir("assets/")))

	// Start the server
	logger.Println("Starting server...")
	if err := http.ListenAndServe(":0", nil); err != nil {
		logger.Fatalln(err)
	}
}
