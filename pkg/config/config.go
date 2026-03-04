package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Mode string

const (
	ModeSimulation Mode = "simulation"
	ModePassive    Mode = "passive"
	ModeOff        Mode = "field_only"
)

type IOBit struct {
	Tag      string `yaml:"tag"`
	Bit      int    `yaml:"bit"`
	Inverted bool   `yaml:"inverted"`
	Type     string `yaml:"type"`
}

type ValveIO struct {
	OpenCmd  IOBit `yaml:"open_cmd"`
	CloseCmd IOBit `yaml:"close_cmd"`
	StopCmd  IOBit `yaml:"stop_cmd"`
	OpenedFB IOBit `yaml:"opened_fb"`
	ClosedFB IOBit `yaml:"closed_fb"`
	Ready    IOBit `yaml:"ready"`
}

type DeviceConfig struct {
	Name         string  `yaml:"name"`
	Type         string  `yaml:"type"`
	OpenTimeSec  int     `yaml:"open_time_sec"`
	CloseTimeSec int     `yaml:"close_time_sec"`
	IO           ValveIO `yaml:"io"`
}

type SCADAConfig struct {
	Commands map[string]IOBit `yaml:"commands"`
}

type PLCConfig struct {
	Path        string `yaml:"path"`
	CycleTimeMs int    `yaml:"cycle_time_ms"`
}

type Config struct {
	Mode         Mode           `yaml:"mode"`
	PLC          PLCConfig      `yaml:"plc"`
	FieldDevices []DeviceConfig `yaml:"field_devices"`
	SCADA        SCADAConfig    `yaml:"scada"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Mode == "" {
		cfg.Mode = ModeOff
	}
	if cfg.PLC.CycleTimeMs <= 0 {
		cfg.PLC.CycleTimeMs = 100
	}

	return &cfg, nil
}

func (c *Config) CycleTime() time.Duration {
	return time.Duration(c.PLC.CycleTimeMs) * time.Millisecond
}
