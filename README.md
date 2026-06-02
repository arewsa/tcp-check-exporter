# tcp-check-exporter

TCP port availability exporter for Prometheus.

`tcp-check-exporter` periodically opens TCP connections to configured hosts and ports, then exposes the result as Prometheus metrics. It is useful for simple availability checks of databases, APIs, load balancers, DNS servers, internal services, and other TCP endpoints.

## Features

- TCP connectivity checks for multiple targets
- Prometheus metrics endpoint
- Per-target and global probe timeouts
- Configurable probe interval and worker count
- Text or JSON logging
- Health endpoint for service checks

## Metrics

| Metric | Type | Labels | Description |
| --- | --- | --- | --- |
| `host_up` | gauge | `host`, `port`, `name` | TCP endpoint status: `1` means available, `0` means unavailable. |
| `host_probe_duration_seconds` | gauge | `host`, `port` | Duration of the last TCP probe. |
| `exporter_scrapes_total` | counter | - | Total number of probe cycles. |
| `exporter_scrape_errors_total` | counter | - | Total number of failed target checks. |
| `exporter_last_scrape_timestamp` | gauge | - | Unix timestamp of the last probe cycle. |

## Quick Start

Clone the repository and prepare the configuration:

```bash
git clone https://github.com/arewsa/tcp-check-exporter.git
cd tcp-check-exporter
cp config-example.yml config.yml
```

Edit `config.yml`, then run the exporter:

```bash
go run .
```

By default the exporter listens on port `8080` and exposes metrics at:

```text
http://localhost:8080/metrics
```

The health endpoint is available at:

```text
http://localhost:8080/health
```

## Configuration

The exporter reads configuration from `config.yml` by default. You can pass another file with `--config.file`.

Minimal example:

```yaml
server:
  listen_address: 8080

targets:
  - host: "8.8.8.8"
    port: 53
    name: "google_dns"
```

Fuller example:

```yaml
server:
  listen_address: 8080
  metrics_path: "metrics"

log:
  level: "info"
  format: "text"
  output: "stdout"

probe:
  timeout: "3s"
  interval: "60s"
  workers: 10

targets:
  - host: "github.com"
    port: 443
    name: "github_https"

  - host: "slow-server.internal"
    port: 8080
    name: "slow_api"
    timeout: "10s"
```

### Server

| Field | Default | Description |
| --- | --- | --- |
| `listen_address` | `8080` | HTTP server port. |
| `metrics_path` | `metrics` | Metrics endpoint path without a leading slash. |

### Logging

| Field | Default | Description |
| --- | --- | --- |
| `level` | `info` | Log level: `debug`, `info`, `warn`, or `error`. |
| `format` | `text` | Log format: `text` or `json`. |
| `output` | `stdout` | Log output: `stdout`, `stderr`, or `file`. |
| `file` | `/var/log/host-exporter.log` | Log file path when `output` is `file`. |

### Probe

| Field | Default | Description |
| --- | --- | --- |
| `timeout` | `3s` | Global TCP connection timeout. |
| `interval` | `60s` | Delay between probe cycles. |
| `workers` | `5` | Number of parallel probe workers. |

Durations use Go duration format, for example `500ms`, `3s`, `1m`.

### Targets

| Field | Required | Description |
| --- | --- | --- |
| `host` | yes | Hostname or IP address to check. |
| `port` | yes | TCP port to connect to. |
| `name` | no | Human-readable target name used in the `host_up` metric label. |
| `timeout` | no | Target-specific timeout. Overrides `probe.timeout`. |

## Command-line Flags

Configuration file values can be overridden with command-line flags:

```bash
go run . \
  --config.file=config.yml \
  --web.listen-address=9090 \
  --web.telemetry-path=metrics \
  --probe.timeout=5s \
  --probe.interval=30s \
  --probe.workers=20 \
  --log.level=debug \
  --log.format=json
```

Available flags:

| Flag | Default | Description |
| --- | --- | --- |
| `--config.file` | `config.yml` | Path to the configuration file. |
| `--web.listen-address` | config value | HTTP server port. |
| `--web.telemetry-path` | config value | Metrics endpoint path without a leading slash. |
| `--probe.timeout` | config value | Global probe timeout. |
| `--probe.interval` | config value | Probe interval. |
| `--probe.workers` | config value | Number of parallel workers. |
| `--log.level` | config value | Log level. |
| `--log.format` | config value | Log format. |
| `--log.output` | config value | Log output destination. |

## Prometheus Scrape Configuration

Example Prometheus job:

```yaml
scrape_configs:
  - job_name: "tcp-check-exporter"
    static_configs:
      - targets:
          - "localhost:8080"
```

If you change `server.metrics_path`, also set `metrics_path` in the Prometheus job:

```yaml
scrape_configs:
  - job_name: "tcp-check-exporter"
    metrics_path: "/probe"
    static_configs:
      - targets:
          - "localhost:8080"
```

## Build

Build a local binary:

```bash
go build -o tcp-check-exporter .
```

Run it:

```bash
./tcp-check-exporter --config.file=config.yml
```
