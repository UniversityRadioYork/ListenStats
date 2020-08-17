package main

import (
	"fmt"
	"github.com/burntsushi/toml"
	"github.com/gorilla/mux"
	"io"
	cfg "listenstats/config"
	"listenstats/handlers"
	"listenstats/reporters"
	"log"
	"net/http"
	"os"
)

func main() {
	var config cfg.Config
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatal(fmt.Errorf("couldn't parse config: %w", err))
	}

	if config.Logfile != "" {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		logfile, err := os.OpenFile(config.Logfile, os.O_APPEND|os.O_CREATE, os.ModeAppend)
		if err != nil {
			log.Fatal(fmt.Errorf("couldn't open logfiel %s: %w", config.Logfile, err))
		}
		log.SetOutput(io.MultiWriter(os.Stdout, logfile))
	}

	var reporter reporters.ListenReporter
	switch config.Reporter {
	case "log":
		reporter = &reporters.LogReporter{}
	default:
		log.Fatal(fmt.Errorf("unknown listener reporter %s", config.Reporter))
	}

	r := mux.NewRouter()

	httpHandler := handlers.NewHttpHandler(&config, reporter)

	r.HandleFunc("/{endpoint}", httpHandler.Handle)

	log.Printf("Listening on %s\n", config.HttpListenAddr)
	log.Fatal(http.ListenAndServe(config.HttpListenAddr, r))
}
