package main

import (
	"log"
	"net/http"

	instana "github.com/instana/go-sensor"
)

var sensor = instana.NewSensor("http-server")

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	if err := http.ListenAndServe(":0", nil); err != nil {
		log.Fatalln(err)
	}
}
