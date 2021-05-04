package main

import (
	"log"
	"net/http"
	"os"

	instana "github.com/instana/go-sensor"
)

var (
	logger = log.New(os.Stdout)
	sensor *instana.Sensor
)

func main() {
	sensor = instana.NewSensor()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	logger.Println("starting server")
	if err := http.ListenAndServe(":0", nil); err != nil {
		log.Fatalln(err)
	}
}
