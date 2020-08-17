package config

type Config struct {
	HttpListenAddr string
	HttpServers    []HttpServer
	TrustedProxies []string
	Reporter       string
	Logfile        string
}

type HttpServer struct {
	BaseUrl      string
	AllowedPaths []string
}
