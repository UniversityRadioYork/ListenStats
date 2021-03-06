package utils

import (
	"listenstats/config"
	"net"
	"net/http"
	"strings"
)

func IsLastProxyTrusted(cfg *config.Config, r *http.Request) bool {
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		panic(err)
	}
	for _, trusted := range cfg.TrustedProxies {
		var network *net.IPNet
		if strings.Contains(trusted, "/") {
			_, network, err = net.ParseCIDR(trusted)
			if err != nil {
				panic(err) // panic is appropriate, since this is user-provided
			}
		} else {
			ip := net.ParseIP(trusted)
			if ip.To4() != nil {
				// ipv4
				network = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
			} else {
				// ipv6
				network = &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
			}
		}
		if network.Contains(net.ParseIP(remoteIp)) {
			return true
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
		if IsLastProxyTrusted(cfg, r) {
			ips := strings.Split(xff, ",")
			addr = strings.TrimSpace(ips[0])
		}
	}
	return addr
}
