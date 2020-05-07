<div align="center">
  <img src="https://cdn.eazyauction.de/eastatic/scx_logo.png">
</div>

# GoProMetrics

The use case for GoProMetrics is to provide an aggregator and metrics cache for ephemeral processes. Such a scripting 
languages like PHP. GoProMetrics is simple, lightweight, fast and provide easy to use API over HTTP.

## Features

* Support for counter using namespace and label
* Histogram support using namespace, label and buckets
* Summary support using namespace, label and objectives
* Provide `/metrics` endpoint for scrape Metrics

## Install - Build it

````
go build

// run it
./goprometrics
````

### Run it

````
./goprometrics -h
Usage of ./goprometrics:
  -host string
        Api Host (default "127.0.0.1")
  -hostm string
        Host to expose metrics (default "127.0.0.1")
  -port string
        Api Port (default "9111")
  -portm string
        Port to expose metrics (default "9112")
````

## Install - Run with Docker

````
docker pull jtlsoftware/goprometrics
docker run -it -p 9111:9111 -p 9112:9112 jtlsoftware/goprometrics
````

## Install - Run with Docker-Compose

````
docker-compose up -d
````

Will start GoProMetrics listen in :9111 (api) and :9112 (for expose metrics). There is also a prometheus up and running on :9090.

Need some logs? `docker-compose logs -f goprometrics`

# Examples

````
curl -XPUT '127.0.0.1:9111/count/foobar/drinks' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'labels=alcoholic:beer'

````
See example directory for Example requests


