package reporters

import (
	"listenstats/config"
	"listenstats/utils"
	"net/http"
	"net/url"
	"time"
)

type ListenerInfo struct {
	IP          string
	QueryParams url.Values
	ServerURL   *url.URL
	Headers     http.Header
}

func MakeListenerInfoFromRequest(cfg *config.Config, proxyUrl *url.URL, r *http.Request) (*ListenerInfo, error) {
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	return &ListenerInfo{
		IP:          utils.FindClientRemoteAddr(cfg, r),
		QueryParams: q,
		ServerURL:   proxyUrl,
		Headers:     r.Header,
	}, nil
}

type ListenReporter interface {
	ReportListenStart(clientId string, info *ListenerInfo) error
	ReportGeoIP(clientId string, info *ListenerInfo)
	ReportListenEnd(clientId string, time time.Duration) error
}
