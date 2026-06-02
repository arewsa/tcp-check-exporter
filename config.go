package main

import (
    "fmt"
    "os"
    "strconv"
    "time"

    "gopkg.in/yaml.v3"
)

type ServerConfig struct {
    ListenAddress int    `yaml:"listen_address"`
    MetricsPath   string `yaml:"metrics_path"`
}

type LogConfig struct {
    Level  string `yaml:"level"`
    Format string `yaml:"format"`
    Output string `yaml:"output"`
    File   string `yaml:"file"`
}

type ProbeConfig struct {
    Timeout  string `yaml:"timeout"`
    Interval string `yaml:"interval"`
    Workers  int    `yaml:"workers"`
}

type TargetConfig struct {
    Host    string `yaml:"host"`
    Port    int    `yaml:"port"`
    Name    string `yaml:"name"`
    Timeout string `yaml:"timeout"`
}

type Config struct {
    Server  ServerConfig   `yaml:"server"`
    Log     LogConfig      `yaml:"log"`
    Probe   ProbeConfig    `yaml:"probe"`
    Targets []TargetConfig `yaml:"targets"`
}

type ParsedConfig struct {
    ListenAddress string
    MetricsPath   string
    LogLevel      string
    LogFormat     string
    LogOutput     string
    LogFile       string
    Timeout       time.Duration
    Interval      time.Duration
    Workers       int
    Targets       []TargetConfig
}

func LoadConfig(filename string) (*ParsedConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("cannot read config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("cannot parse YAML: %w", err)
    }

    parsed := &ParsedConfig{
        ListenAddress: ":" + strconv.Itoa(cfg.Server.ListenAddress),
        MetricsPath:   cfg.Server.MetricsPath,
        LogLevel:      cfg.Log.Level,
        LogFormat:     cfg.Log.Format,
        LogOutput:     cfg.Log.Output,
        LogFile:       cfg.Log.File,
        Workers:       cfg.Probe.Workers,
        Targets:       cfg.Targets,
    }

    if cfg.Server.ListenAddress == 0 {
        parsed.ListenAddress = ":8080"
    }
    if parsed.MetricsPath == "" {
        parsed.MetricsPath = "metrics"
    }
    if parsed.LogLevel == "" {
        parsed.LogLevel = "info"
    }
    if parsed.LogFormat == "" {
        parsed.LogFormat = "text"
    }
    if parsed.LogOutput == "" {
        parsed.LogOutput = "stdout"
    }
    if parsed.Workers <= 0 {
        parsed.Workers = 5
    }

    timeout, err := time.ParseDuration(cfg.Probe.Timeout)
    if err != nil {
        timeout = 3 * time.Second
    }
    parsed.Timeout = timeout

    interval, err := time.ParseDuration(cfg.Probe.Interval)
    if err != nil {
        interval = 60 * time.Second
    }
    parsed.Interval = interval

    return parsed, nil
}

func (c *ParsedConfig) GetTimeoutForTarget(target TargetConfig) time.Duration {
    if target.Timeout != "" {
        if timeout, err := time.ParseDuration(target.Timeout); err == nil {
            return timeout
        }
    }
    return c.Timeout
}