<div align="center">
  <img src="https://cdn.eazyauction.de/eastatic/scx_logo.png">
</div>

# GoProMetrics

![Test](https://github.com/jtl-software/goprometrics/workflows/Test/badge.svg?branch=master)

The use case for GoProMetrics is to provide an aggregator and metrics cache for ephemeral processes. Such a scripting 
languages like PHP. GoProMetrics is simple, lightweight, fast and provide easy to use API over HTTP.

## Features

* Support for Counter metrics
* Support for Gauge metrics
* Support for Histogram metrics
* Support form Summary metrics
* Each metric can be described using Namespace, Labels and Help Text
* Provide `/metrics` endpoint for scrapping

## Install - Build it

````
go build

// run it
./goprometrics
````

## Install - Run with Docker

````
docker pull jtlsoftware/goprometrics
docker run -it -p 9111:9111 -p 9112:9112 jtlsoftware/goprometrics
````

## Install - Using Docker-Compose

````
docker-compose up -d
````

Will start GoProMetrics listen in :9111 (api) and :9112 (for expose metrics). There is also a prometheus up and running on :9090.

Need some logs? `docker-compose logs -f goprometrics`

# Examples

Push and increment a counter
````
curl -XPUT '127.0.0.1:9111/count/foobar/drinks' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'labels=alcoholic:beer'
````

Expose Metrics
````
curl -XPUT '127.0.0.1:9112/metrics'
````

See example directory for Example requests


