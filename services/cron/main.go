package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"encoding/json"

	"github.com/braintree/manners"
	"github.com/robfig/cron"
)

type Request struct {
	CronSettings string `json:"cron_settings"`
	Endpoint     string `json:"endpoint"`
}

func makeRequest(url string) func() {
	return func() {
		_, err := http.Get(url)
		if err != nil {
			log.Println("Request error")
		} else {
			log.Println("Request sended")
		}
	}
}

func main() {
	var httpAddr = flag.String("http", "0.0.0.0:8000", "HTTP service address")
	flag.Parse()

	c := cron.New()
	c.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var request Request
		err := decoder.Decode(&request)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()
		c.Stop()
		c.AddFunc(request.CronSettings, makeRequest(request.Endpoint))
		c.Start()

	})

	log.Println("Starting server ...")
	log.Printf("HTTP service listening on %s", *httpAddr)

	httpServer := manners.NewServer()
	httpServer.Addr = *httpAddr
	httpServer.Handler = mux

	errChan := make(chan error, 10)

	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatal(err)
			}
		case s := <-signalChan:
			log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			httpServer.BlockingClose()
			os.Exit(0)
		}
	}
}
