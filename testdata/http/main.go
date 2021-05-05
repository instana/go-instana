package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	client := http.Client{}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		resp, err := client.Get("https://wttr.in")
		if err != nil {
			logger.Println("failed to send request:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		defer resp.Body.Close()

		w.WriteHeader("Content-Type", "text/plain")
		io.Copy(w, resp.Body)
	})

	http.Handle("/assets", http.FileServer(http.Dir("assets/")))

	logger.Println("Starting server...")
	if err := http.ListenAndServe(":0", nil); err != nil {
		logger.Fatalln(err)
	}
}
