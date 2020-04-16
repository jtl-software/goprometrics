<div align="center">
  <img src="https://cdn.eazyauction.de/eastatic/scx_logo.png">
</div>

# goprometrics

The use case for Goprometheus is to provide an aggregator and metrics cache for ephemeral processes. Such a scripting 
languages like PHP. Goprometheus is simple, lightweight and fast and provide a easy to use API over HTTP.

## Features

* Support for counter using namespace and label
* Histogram support using namespace, label and buckets
* Summary support using namespace, label and objectives
* Provide `/metrics` endpoint for scrape Metrics

## Install - Build it

````
go build
````

## Install - using docker

````
docker build --tag goprometrics:0.1 ./
docker run -it -p 9111:9111 -p 9112:9112 -v $PWD/src:/go/src/goprometrics goprometrics:0.1
````

Docker container goprometrics will host the metric collector and running using `refresh`, which means every source change
will trigger a automatic re-build. You may not is it this way in your production.

## Install - using docker-compose

````
docker-compose up -d
````

Will start goprometrics listen in :9111 (api) and :9112 (for expose metrics). There is also a prometheus up and running on :9090.

Need some logs? `docker-compose logs -f goprometrics`

# Examples

See example.http for Example requests


