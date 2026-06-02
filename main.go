// main.go
package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := LoadConfig("config.yml")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	log.Printf("Loaded %d targets", len(cfg.Targets))
	log.Printf("Server: %d, Metrics: %s", cfg.ListenPort, cfg.MetricsPath)
	log.Printf("Global timeout: %v, Interval: %v", cfg.Timeout, cfg.Interval)

	metrics := NewMetrics()

	go func() {
		for {
			metrics.ScrapesTotal.Inc()

			startTime := time.Now()

			log.Println("Starting scrape cycle...")

			errorsInCycle := 0

			for _, target := range cfg.Targets {
				timeout := cfg.Timeout

				up, latency := checkHost(target.Host, target.Port, timeout)

				if up {
					metrics.HostUp.WithLabelValues(target.Host, strconv.Itoa(target.Port), target.Name).Set(1)
					log.Printf("%s:%d (%s) - OK (%.3fs)",
						target.Host, target.Port, target.Name, latency)
				} else {
					metrics.HostUp.WithLabelValues(target.Host, strconv.Itoa(target.Port), target.Name).Set(0)
					log.Printf("%s:%d (%s) - FAIL (%.3fs)",
						target.Host, target.Port, target.Name, latency)
					errorsInCycle++
				}

				metrics.HostLatency.WithLabelValues(target.Host, strconv.Itoa(target.Port)).Set(latency)
			}

			if errorsInCycle > 0 {
				metrics.ScrapeErrorsTotal.Add(float64(errorsInCycle))
			}

			metrics.LastScrapeTimestamp.Set(float64(startTime.Unix()))

			log.Printf("Scrape cycle completed, waiting %v...", cfg.Interval)
			time.Sleep(cfg.Interval)
		}
	}()

	http.Handle("/"+cfg.MetricsPath, promhttp.Handler())

	log.Printf("Starting HTTP server on %d", cfg.ListenPort)
	log.Printf("Metrics available at http://localhost:%d/%s",
		cfg.ListenPort, cfg.MetricsPath)

	if err := http.ListenAndServe(":"+strconv.Itoa(cfg.ListenPort), nil); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}
