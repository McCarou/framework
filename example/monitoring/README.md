# Example application for the Radian framework

## Usage

#### 1 Clone the repository

```
git clone https://github.com/radianteam/framework.git
```
```
cd framework
```

#### 2 Goto this folder

```
cd example/monitoring
```


#### 3 Run the application

```
docker-compose up -d
```

#### 4 Make some requests
Commands:
```
curl 127.0.0.1:8088/ 
```
```
curl 127.0.0.1:8088/absent
```

Example
```
radian@radian:~$ curl 127.0.0.1:8088/                                   
Hello world!
radian@radian:~$ curl 127.0.0.1:8088/absent
404 page not found
```

#### 5 Check metrics
Commands:
```
curl 127.0.0.1:8087/metrics
```

Example
```
radian@radian:~$ curl 127.0.0.1:8087/metrics
...
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP rest_worker_total_requests Total requests of the rest worker
# TYPE rest_worker_total_requests counter
rest_worker_total_requests{code="200",method="GET",url="/",worker_name="service_rest"} 1
rest_worker_total_requests{code="404",method="GET",url="/absent",worker_name="service_rest"} 1
```

#### 6 Enjoy!

And don't forget to stop the application :)

```
docker-compose down
```
