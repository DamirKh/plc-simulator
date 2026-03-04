package scada

import (
	"fmt"
	"time"

	"plc-simulator/pkg/io"
)

type Mode string

const (
	ModeSimulation Mode = "simulation"
	ModePassive    Mode = "passive"
	ModeOff        Mode = "field_only"
)

type Simulator struct {
	mode     Mode
	Commands map[string]io.DiscreteWriter
}

func NewSimulator(mode Mode) *Simulator {
	return &Simulator{
		mode:     mode,
		Commands: make(map[string]io.DiscreteWriter),
	}
}

func (s *Simulator) SetMode(mode Mode) {
	s.mode = mode
}

func (s *Simulator) Mode() Mode {
	return s.mode
}

func (s *Simulator) WriteCommand(name string, value bool) error {
	if s.mode != ModeSimulation {
		return fmt.Errorf("mode=%s: команды только в simulation", s.mode)
	}

	cmd, ok := s.Commands[name]
	if !ok {
		return fmt.Errorf("команда не найдена: %s", name)
	}

	return cmd.Write(value)
}

func (s *Simulator) PulseCommand(name string, duration time.Duration) error {
	if err := s.WriteCommand(name, true); err != nil {
		return err
	}
	time.Sleep(duration)
	return s.WriteCommand(name, false)
}
