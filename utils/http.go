package utils

import (
	"listenstats/config"
	"net"
	"net/http"
	"strings"
)

func IsLastProxyTrusted(cfg *config.Config, xffValue string, r *http.Request) bool {
	ips := strings.Split(xffValue, ", ")
	lastProxy := ips[len(ips)-1]
	for _, trusted := range cfg.TrustedProxies {
		if lastProxy == trusted {
			// We trust the last proxy, assume they know what the right IP is
			// First though, check that the request really did come from them
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				if ip == trusted {
					return true
				}
			}
		}
	}
	return false
}

func FindClientRemoteAddr(cfg *config.Config, r *http.Request) string {
	// Start with what they say it is
	addr := strings.Split(r.RemoteAddr, ":")[0]
	// Then check XFF
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Check the identity of the last proxy to touch it
		// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
		if IsLastProxyTrusted(cfg, xff, r) {
			ips := strings.Split(xff, ", ")
			addr = ips[0]
		}
	}
	return addr
}
