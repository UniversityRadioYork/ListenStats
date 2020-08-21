package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"io"
	"listenstats/config"
	"listenstats/reporters"
	"listenstats/utils"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

type HttpHandler struct {
	cfg          *config.Config
	reporter     reporters.ListenReporter
	reverseProxy *httputil.ReverseProxy
}

var didLocalIpWarn = false

func NewHttpHandler(cfg *config.Config, reporter reporters.ListenReporter) *HttpHandler {
	var def *config.HttpServer
	for i, serv := range cfg.HttpServers {
		if serv.Default {
			def = &cfg.HttpServers[i]
		}
	}
	handler := &HttpHandler{cfg: cfg, reporter: reporter}
	if def != nil {
		defUrl, err := url.Parse(def.BaseUrl)
		if err != nil {
			panic(err)
		}
		director := func(req *http.Request) {
			req.URL.Scheme = defUrl.Scheme
			req.URL.Host = defUrl.Host
			req.Host = defUrl.Host
			req.URL.Path = singleJoiningSlash(defUrl.Path, req.URL.Path)
			if defUrl.RawQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = defUrl.RawQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = defUrl.RawQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
			var localIp string
			if cfg.HttpLocalIp != "" {
				if !didLocalIpWarn {
					log.Println("No local IP set in config, using default")
					didLocalIpWarn = true
				}
				localIp = "127.0.0.1"
			} else {
				localIp = cfg.HttpLocalIp
			}
			xff := req.Header.Get("X-Forwarded-For")
			if xff != "" {
				if utils.IsLastProxyTrusted(cfg, req) {
					req.Header.Set("X-Forwarded-For", xff+", "+localIp)
				} else {
					// Untrusted proxy; reset XFF
					req.Header.Set("X-Forwarded-For", "")
				}
			} else {
				ip, _, err := net.SplitHostPort(req.RemoteAddr)
				if err != nil {
					panic(err)
				}
				req.Header.Set("X-Forwarded-For", ip+", "+localIp)
			}
		}
		handler.reverseProxy = &httputil.ReverseProxy{Director: director}
	}
	return handler
}

func (h *HttpHandler) Handle(w http.ResponseWriter, r *http.Request) {
	requestId := xid.New()
	w.Header().Add("X-URY-RequestID", requestId.String())
	vars := mux.Vars(r)

	var server *config.HttpServer
	for _, srv := range h.cfg.HttpServers {
		for _, path := range srv.AllowedPaths {
			if "/"+vars["endpoint"] == path {
				server = &srv
			}
		}
	}
	if server == nil {
		if h.reverseProxy != nil {
			// let the reverse proxy handle it
			log.Printf("[%s] path (%s) non-match, passing to default\n", requestId, r.URL.Path)
			h.reverseProxy.ServeHTTP(w, r)
			return
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	if r.Method != "GET" {
		log.Printf("[%s] invalid method %s", requestId, r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reverseUrl, err := url.Parse(server.BaseUrl)
	if err != nil {
		log.Println(fmt.Errorf("couldn't parse server URL [%s]: %w", server.BaseUrl, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	reverseUrl.Path = singleJoiningSlash(reverseUrl.Path, vars["endpoint"])

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

	defer reverseRes.Body.Close()

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
