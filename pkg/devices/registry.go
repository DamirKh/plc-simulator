package devices

import (
	"context"
	"fmt"
	"time"
)

// Device интерфейс для всех устройств
type Device interface {
	Update(dt time.Duration) error
	String() string
}

// Registry хранит все устройства и управляет циклом
type Registry struct {
	devices []Device
}

func NewRegistry() *Registry {
	return &Registry{
		devices: make([]Device, 0),
	}
}

func (r *Registry) Add(d Device) {
	r.devices = append(r.devices, d)
}

func (r *Registry) Run(ctx context.Context, cycleTime time.Duration) error {
	ticker := time.NewTicker(cycleTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for _, d := range r.devices {
				if err := d.Update(cycleTime); err != nil {
					return fmt.Errorf("%s: %w", d, err)
				}
			}
		}
	}
}

func (r *Registry) Status() string {
	result := ""
	for i, d := range r.devices {
		if i > 0 {
			result += " | "
		}
		result += d.String()
	}
	return result
}
