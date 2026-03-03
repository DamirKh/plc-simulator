package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// IOBit описывает один дискретный сигнал
type IOBit struct {
	Tag      string `yaml:"tag"`
	Bit      int    `yaml:"bit"`
	Inverted bool   `yaml:"inverted"`
	Type     string `yaml:"type"` // "input" или "output" (с точки зрения PLC!)
}

// ValveIO конфигурация IO задвижки
type ValveIO struct {
	OpenCmd  IOBit `yaml:"open_cmd"`
	CloseCmd IOBit `yaml:"close_cmd"`
	StopCmd  IOBit `yaml:"stop_cmd"`
	OpenedFB IOBit `yaml:"opened_fb"`
	ClosedFB IOBit `yaml:"closed_fb"`
	Ready    IOBit `yaml:"ready"`
}

// DeviceConfig универсальная конфигурация устройства
type DeviceConfig struct {
	Name         string  `yaml:"name"`
	Type         string  `yaml:"type"`
	OpenTimeSec  int     `yaml:"open_time_sec"`
	CloseTimeSec int     `yaml:"close_time_sec"`
	IO           ValveIO `yaml:"io"`
}

// PLCConfig конфигурация подключения
type PLCConfig struct {
	Path        string `yaml:"path"`
	CycleTimeMs int    `yaml:"cycle_time_ms"`
}

// Config корневой конфиг
type Config struct {
	PLC     PLCConfig      `yaml:"plc"`
	Devices []DeviceConfig `yaml:"devices"`
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

	// Валидация
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.PLC.Path == "" {
		return fmt.Errorf("plc.path is required")
	}
	if c.PLC.CycleTimeMs <= 0 {
		c.PLC.CycleTimeMs = 100 // default
	}
	for i, dev := range c.Devices {
		if dev.Name == "" {
			return fmt.Errorf("device[%d]: name is required", i)
		}
		if dev.Type != "motorized_valve" {
			return fmt.Errorf("device[%d]: unknown type %q", i, dev.Type)
		}
		// Проверяем IO
		if err := validateIOBit(dev.IO.OpenCmd, "open_cmd"); err != nil {
			return fmt.Errorf("device %s: %w", dev.Name, err)
		}
		if err := validateIOBit(dev.IO.CloseCmd, "close_cmd"); err != nil {
			return fmt.Errorf("device %s: %w", dev.Name, err)
		}
		// ... и так для всех сигналов
	}
	return nil
}

func validateIOBit(io IOBit, name string) error {
	if io.Tag == "" {
		return fmt.Errorf("%s: tag is required", name)
	}
	if io.Type != "input" && io.Type != "output" {
		return fmt.Errorf("%s: type must be 'input' or 'output', got %q", name, io.Type)
	}
	if io.Bit < 0 || io.Bit > 31 {
		return fmt.Errorf("%s: bit must be 0-31, got %d", name, io.Bit)
	}
	return nil
}

func (c *Config) CycleTime() time.Duration {
	return time.Duration(c.PLC.CycleTimeMs) * time.Millisecond
}
