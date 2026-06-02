package main

import (
    "context"
    "flag"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"
	"log/slog"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    Version   = "dev"
    BuildTime = "unknown"
    GoVersion = "unknown"
)

func main() {
    configFile := flag.String("config.file", "config.yml", "Path to configuration file")
    listenAddress := flag.Int("web.listen-address", 0, "Address to listen on")
    metricsPath := flag.String("web.telemetry-path", "", "Path to expose metrics")
    logLevel := flag.String("log.level", "", "Log level (debug, info, warn, error)")
    logFormat := flag.String("log.format", "", "Log format (json, text)")
    logOutput := flag.String("log.output", "", "Log output (stdout, stderr, file)")
    probeTimeout := flag.String("probe.timeout", "", "Probe timeout")
    probeInterval := flag.String("probe.interval", "", "Probe interval")
    probeWorkers := flag.Int("probe.workers", 0, "Number of parallel workers")
    flag.Parse()

    cfg, err := LoadConfig(*configFile)
    if err != nil {
        fmt.Printf("Warning: cannot load config file: %v\n", err)
        cfg = &ParsedConfig{
            ListenAddress: ":8080",
            MetricsPath:   "metrics",
            LogLevel:      "info",
            LogFormat:     "text",
            LogOutput:     "stdout",
            Timeout:       3 * time.Second,
            Interval:      60 * time.Second,
            Workers:       5,
        }
    }

    if *listenAddress != 0 {
        cfg.ListenAddress = ":" + strconv.Itoa(*listenAddress)
    }
    if *metricsPath != "" {
        cfg.MetricsPath = *metricsPath
    }
    if *logLevel != "" {
        cfg.LogLevel = *logLevel
    }
    if *logFormat != "" {
        cfg.LogFormat = *logFormat
    }
    if *logOutput != "" {
        cfg.LogOutput = *logOutput
    }
    if *probeTimeout != "" {
        if timeout, err := time.ParseDuration(*probeTimeout); err == nil {
            cfg.Timeout = timeout
        }
    }
    if *probeInterval != "" {
        if interval, err := time.ParseDuration(*probeInterval); err == nil {
            cfg.Interval = interval
        }
    }
    if *probeWorkers > 0 {
        cfg.Workers = *probeWorkers
    }

    if err := SetupLogging(cfg); err != nil {
        fmt.Printf("Failed to setup logging: %v\n", err)
        os.Exit(1)
    }

    metrics := NewMetrics()

    pool := NewWorkerPool(cfg.Workers)
    pool.Start()

    go func() {
        for {
            metrics.ScrapesTotal.Inc()
            startTime := time.Now()

            slog.Info("Starting scrape cycle",
                "workers", cfg.Workers,
                "targets", len(cfg.Targets),
            )

            for _, target := range cfg.Targets {
                timeout := cfg.GetTimeoutForTarget(target)
                pool.AddTask(ProbeTask{
                    Target:  target,
                    Timeout: timeout,
                })
            }

            resultsCount := 0
            errorsInCycle := 0

            for resultsCount < len(cfg.Targets) {
                result := <-pool.Results()

                if result.Up {
                    metrics.HostUp.WithLabelValues(
                        result.Target.Host,
                        strconv.Itoa(result.Target.Port),
                        result.Target.Name,
                    ).Set(1)

                    slog.Debug("Host check successful",
                        "host", result.Target.Host,
                        "port", result.Target.Port,
                        "name", result.Target.Name,
                        "latency_ms", result.Latency*1000,
                    )
                } else {
                    metrics.HostUp.WithLabelValues(
                        result.Target.Host,
                        strconv.Itoa(result.Target.Port),
                        result.Target.Name,
                    ).Set(0)

                    slog.Warn("Host check failed",
                        "host", result.Target.Host,
                        "port", result.Target.Port,
                        "name", result.Target.Name,
                        "latency_ms", result.Latency*1000,
                    )
                    errorsInCycle++
                }

                metrics.HostLatency.WithLabelValues(
                    result.Target.Host,
                    strconv.Itoa(result.Target.Port),
                ).Set(result.Latency)

                resultsCount++
            }

            if errorsInCycle > 0 {
                metrics.ScrapeErrorsTotal.Add(float64(errorsInCycle))
            }

            metrics.LastScrapeTimestamp.Set(float64(startTime.Unix()))

            slog.Info("Scrape cycle completed",
                "duration_ms", time.Since(startTime).Milliseconds(),
                "errors", errorsInCycle,
                "targets", len(cfg.Targets),
            )

            time.Sleep(cfg.Interval)
        }
    }()

    http.Handle("/"+cfg.MetricsPath, promhttp.Handler())
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    srv := &http.Server{
        Addr:    cfg.ListenAddress,
        Handler: http.DefaultServeMux,
    }

    go func() {
        slog.Info("Starting HTTP server", "address", cfg.ListenAddress)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("HTTP server failed", "error", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
    <-quit

    slog.Info("Shutting down gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("HTTP server shutdown error", "error", err)
    }

    pool.Close()
    slog.Info("Server stopped")
}