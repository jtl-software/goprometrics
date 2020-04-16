package main

import "flag"

type HostConfig struct {
	host string
	port string
}

type Config struct {
	apiHostConfig  HostConfig
	mApiHostConfig HostConfig
}

func NewConfig() Config {

	var host = flag.String("host", "127.0.0.1", "Api Host")
	var port = flag.String("port", "9111", "Api Port")
	var hostMetrics = flag.String("hostm", "127.0.0.1", "Host for prometheus scraping")
	var portMetrics = flag.String("portm", "9112", "Port for prometheus scraping")
	flag.Parse()

	return Config{
		apiHostConfig: HostConfig{
			host: *host,
			port: *port,
		},
		mApiHostConfig: HostConfig{
			host: *hostMetrics,
			port: *portMetrics,
		},
	}
}
