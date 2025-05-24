package metrics

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    PagesCrawled = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "crawler_pages_total",
            Help: "Total pages successfully crawled",
        },
        []string{"task_tag"},
    )

    CrawlFailures = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "crawler_failures_total",
            Help: "Total crawl failures",
        },
        []string{"task_tag"},
    )
)

func Init() {
    prometheus.MustRegister(PagesCrawled, CrawlFailures)
}

func Serve() {
    http.Handle("/metrics", promhttp.Handler())
    go http.ListenAndServe(":2112", nil)
}
