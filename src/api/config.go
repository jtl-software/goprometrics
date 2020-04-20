package api

import "flag"

type HostConfig struct {
	host string
	port string
}

type Config struct {
	ApiHostConfig        HostConfig
	MetricsApiHostConfig HostConfig
}

func NewConfig() Config {

	var host = flag.String("host", "", "Api Host")
	var port = flag.String("port", "9111", "Api Port")
	var hostMetrics = flag.String("hostm", "", "Host for prometheus scraping")
	var portMetrics = flag.String("portm", "9112", "Port for prometheus scraping")
	flag.Parse()

	return Config{
		ApiHostConfig: HostConfig{
			host: *host,
			port: *port,
		},
		MetricsApiHostConfig: HostConfig{
			host: *hostMetrics,
			port: *portMetrics,
		},
	}
}
