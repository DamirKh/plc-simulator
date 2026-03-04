package field

import (
	"context"
	"fmt"
	"time"
)

type Device interface {
	Update(dt time.Duration) error
	String() string
}

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

func (r *Registry) Update(dt time.Duration) error {
	for _, d := range r.devices {
		if err := d.Update(dt); err != nil {
			return fmt.Errorf("%s: %w", d, err)
		}
	}
	return nil
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

func (r *Registry) Run(ctx context.Context, cycleTime time.Duration) error {
	ticker := time.NewTicker(cycleTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.Update(cycleTime); err != nil {
				return err
			}
		}
	}
}
