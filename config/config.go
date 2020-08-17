package config

import "time"

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
	TrustedProxies []string
	Reporter       string
	Logfile        string
	Postgres       PostgresReporter

	HttpServers []HttpServer
}

type HttpServer struct {
	BaseUrl      string
	AllowedPaths []string
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
