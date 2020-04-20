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

## Install - using docker

````
docker build --tag goprometrics:0.1 ./
docker run -it -p 9111:9111 -p 9112:9112 -v $PWD/src:/go/src/goprometrics goprometrics:0.1
````

Docker container goprometrics will host the metric collector and running using `refresh`, which means every source change
will trigger a automatic re-build. You may not use it this way in your production.

## Install - using docker-compose

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


