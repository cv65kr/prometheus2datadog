# Prometheus2Datadog

The tool allows to convert and send metrics from prometheus to datadog (statsd).

Example of prometheus format:
```
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 4.3223e-05
go_gc_duration_seconds{quantile="0.25"} 0.000202335
go_gc_duration_seconds{quantile="0.5"} 0.000329636
go_gc_duration_seconds{quantile="0.75"} 0.000492795
go_gc_duration_seconds{quantile="1"} 0.670895851
go_gc_duration_seconds_sum 1.463038969
go_gc_duration_seconds_count 260
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 48
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.21.2"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 9.898136e+06
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 1.754062264e+09
```

Supported types:
- COUNTER
- GAUGE
- HISTOGRAM
- SUMMARY

## Help
```bash
./prometheus2datadog help
```

```
Usage: ./prometheus2datadog [OPTIONS] argument ...
  -exclude string
        Prefix for excluded metrics e.g. test_,xxx_,www_
  -log-level string
        log level debug, info, warn, error, fatal or panic (default "debug")
  -metrics-endpoint string
        Endpoint for scrapping e.g. 'http://localhost:2021'. Should contains prometheus format metrics (default "http://0.0.0.0:2112")
  -scraping-interval int
        Interval for metrics scraping in seconds (default 60)
  -shutdown-timeout int
        Gracefull shutdown timeout in seconds (default 60)
  -statsd-address string
        Address for statsd e.g. 'localhost:8125' (default "0.0.0.0:8125")
```

## Run
```bash
./prometheus2datadog
```
