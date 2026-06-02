package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	ListenPort  int    `yaml:"listen_port"`
	MetricsPath string `yaml:"metrics_path"`
}

type ProbeConfig struct {
	Timeout  string `yaml:"timeout"`
	Interval string `yaml:"interval"`
}

type TargetConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Name string `yaml:"name"`
}

type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Probe   ProbeConfig    `yaml:"probe"`
	Targets []TargetConfig `yaml:"targets"`
}

type ParsedConfig struct {
	ListenPort  int
	MetricsPath string
	Timeout     time.Duration
	Interval    time.Duration
	Targets     []TargetConfig
}

func LoadConfig(filename string) (*ParsedConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Cannot read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Cannot parse YAML: %w", err)
	}

	timeout, err := time.ParseDuration(cfg.Probe.Timeout)
	if err != nil {
		return nil, fmt.Errorf("Invalid probe.timeout: %w", err)
	}

	interval, err := time.ParseDuration(cfg.Probe.Interval)
	if err != nil {
		return nil, fmt.Errorf("Invalid probe.interval: %w", err)
	}

	parsed := &ParsedConfig{
		ListenPort:  cfg.Server.ListenPort,
		MetricsPath: cfg.Server.MetricsPath,
		Timeout:     timeout,
		Interval:    interval,
		Targets:     cfg.Targets,
	}

	if parsed.ListenPort == 0 {
		parsed.ListenPort = 8080
	}

	if parsed.MetricsPath == "" {
		parsed.MetricsPath = "metrics"
	}

	return parsed, nil
}
