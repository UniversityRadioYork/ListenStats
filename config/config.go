package config

import (
	"fmt"
	"listenstats/utils"
	"time"
)

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type Config struct {
	HttpListenAddr string
	HttpLocalIp    string
	Reporter       string
	Logfile        string
	Postgres       PostgresReporter

	TrustedProxies []string
	TrustedCDNs    []string

	HttpServers []HttpServer
}

func (c *Config) Init() error {
	// For future flexibility
	return c.GetCDNIPs()
}

func (c *Config) GetCDNIPs() error {
	for _, cdn := range c.TrustedCDNs {
		switch cdn {
		case "cloudflare":
			cloudflareIps, err := utils.GetCloudflareIPRanges()
			if err != nil {
				return err
			}
			c.TrustedProxies = append(c.TrustedProxies, cloudflareIps...)
		default:
			return fmt.Errorf("Unknown CDN %s", cdn)
		}
	}
	return nil
}

type HttpServer struct {
	BaseUrl      string
	AllowedPaths []string
	Default      bool
}

type PostgresReporter struct {
	Host          string
	Port          int
	User          string
	Password      string
	Database      string
	Schema        string
	MinListenTime duration
}
