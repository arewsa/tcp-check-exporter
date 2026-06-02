// metrics.go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
    HostUp *prometheus.GaugeVec
    HostLatency *prometheus.GaugeVec
    ScrapesTotal prometheus.Counter
    ScrapeErrorsTotal prometheus.Counter
    LastScrapeTimestamp prometheus.Gauge
}

func NewMetrics() *Metrics {
    hostUp := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "host_up",
            Help: "Host availability status (1 = up, 0 = down)",
        },
        []string{"host", "port", "name"},
    )
    
    hostLatency := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "host_probe_duration_seconds",
            Help: "Probe duration in seconds",
        },
        []string{"host", "port"},
    )
    
    scrapesTotal := prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "exporter_scrapes_total",
            Help: "Total number of scrapes (check cycles)",
        },
    )
    
    scrapeErrorsTotal := prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "exporter_scrape_errors_total",
            Help: "Total number of scrape errors",
        },
    )
    
    lastScrapeTimestamp := prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "exporter_last_scrape_timestamp",
            Help: "Unix timestamp of last successful scrape",
        },
    )
    
    metrics := &Metrics{
        HostUp:              hostUp,
        HostLatency:         hostLatency,
        ScrapesTotal:        scrapesTotal,
        ScrapeErrorsTotal:   scrapeErrorsTotal,
        LastScrapeTimestamp: lastScrapeTimestamp,
    }
    
    prometheus.MustRegister(hostUp)
    prometheus.MustRegister(hostLatency)
    prometheus.MustRegister(scrapesTotal)
    prometheus.MustRegister(scrapeErrorsTotal)
    prometheus.MustRegister(lastScrapeTimestamp)
    
    return metrics
}