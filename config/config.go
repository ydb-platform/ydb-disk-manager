package config

import "time"

type Config struct {
	DiskMatch      string        `yaml:"diskMatch"`
	HostProcPath   string        `yaml:"hostProcPath"`
	UpdateInterval time.Duration `yaml:"updateInterval"`
	DeviceCount    uint          `yaml:"deviceCount"`
}
