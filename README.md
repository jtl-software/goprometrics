# JTLPROM

````
docker-compose up -d
````

This will start jtlprom listen in :9111 (api) and :9112 (for expose metrics). There is also a prometheus up and running
on :9090

See example.http for Example requests

Docker container jtlprom will host the metric collector and running with `refrsh`, which means every source change will
trigger a automatic re-build. 

See `docker-compose logs -f jtlprom` for logs
