package plc

import "time"

// Metrics structure to hold the performance metrics
type Metrics struct {
	LastUpdaterDuration time.Duration
	UpdateIntervalUtilization float64 // Added field for utilization
}

// Cache structure to hold cache data
type Cache struct {
	updateInterval time.Duration // Added field for update interval
}

// NewCache function to initialize a Cache
func NewCache(updateInterval time.Duration) *Cache {
	return &Cache{
		updateInterval: updateInterval, // Store updateInterval
	}
}

// GetMetrics method to calculate utilization
func (c *Cache) GetMetrics() Metrics {
	utilization := float64(c.LastUpdaterDuration) / float64(c.updateInterval)
	if utilization > 1.0 {
		utilization = 1.0 // Clamp to 1.0
	}
	return Metrics{
		LastUpdaterDuration: c.LastUpdaterDuration,
		UpdateIntervalUtilization: utilization,
	}
}