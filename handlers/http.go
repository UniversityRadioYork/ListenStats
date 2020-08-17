package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"io"
	"listenstats/config"
	"listenstats/reporters"
	"log"
	"net/http"
	"net/url"
	"time"
)

type HttpHandler struct {
	cfg      *config.Config
	reporter reporters.ListenReporter
}

func NewHttpHandler(cfg *config.Config, reporter reporters.ListenReporter) *HttpHandler {
	return &HttpHandler{cfg: cfg, reporter: reporter}
}

func (h *HttpHandler) Handle(w http.ResponseWriter, r *http.Request) {
	requestId := xid.New()
	w.Header().Add("X-URY-RequestID", requestId.String())
	vars := mux.Vars(r)

	if r.Method != "GET" {
		log.Printf("[%s] invalid method %s", requestId, r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var server *config.HttpServer
	for _, srv := range h.cfg.HttpServers {
		for _, path := range srv.AllowedPaths {
			if "/"+vars["endpoint"] == path {
				server = &srv
			}
		}
	}
	if server == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reverseUrl, err := url.Parse(server.BaseUrl)
	if err != nil {
		log.Println(fmt.Errorf("couldn't parse server URL [%s]: %w", server.BaseUrl, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	reverseUrl.Path = "/" + vars["endpoint"]

	listenerInfo, err := reporters.MakeListenerInfoFromRequest(h.cfg, reverseUrl, r)
	if err != nil {
		log.Println(fmt.Errorf("[%s] couldn't make listener info %T: %w", requestId, h.reporter, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("[%s] @ %s for %s hello\n", requestId, listenerInfo.IP, reverseUrl.String())

	reportStart := time.Now()
	if err := h.reporter.ReportListenStart(requestId.String(), listenerInfo); err != nil {
		log.Println(fmt.Errorf("[%s] couldn't report listen start to reporter %T: %w", requestId, h.reporter, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	reportEnd := time.Now()
	log.Printf("[%s] reporting to %T took %v\n", requestId, h.reporter, reportEnd.Sub(reportStart))

	reverseRes, err := http.Get(reverseUrl.String())
	if err != nil {
		log.Println(fmt.Errorf("[%s] reverse proxy request failed: %w", requestId, err))
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if reverseRes.StatusCode != 200 {
		log.Printf("[%s] reverse proxy non-200 (%s), passing through\n", requestId, reverseRes.Status)
	}
	for key, vals := range reverseRes.Header {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
	w.WriteHeader(reverseRes.StatusCode)

	timeStart := time.Now()
	_, err = io.Copy(w, reverseRes.Body)
	timeEnd := time.Now()

	alive := timeEnd.Sub(timeStart)
	if err != nil {
		log.Println(fmt.Errorf("[%s] reverse proxy passthrough failed: %w", requestId, err))
	}
	log.Printf("[%s] goodbye, was nice knowing you for %v\n", requestId, alive)
	if err := h.reporter.ReportListenEnd(requestId.String(), alive); err != nil {
		log.Println(fmt.Errorf("[%s] couldn't report listen end to reporter %T: %w", requestId, h.reporter, err))
	}
}
