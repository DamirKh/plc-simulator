package plc

import (
	// "context"
	"sync"
	"time"
)

// Cache хранит значения тегов с фоновым обновлением
type Cache struct {
	client *Client
	data   map[string]interface{}
	mu     sync.RWMutex
	tags   map[string]struct{}
}

// NewCache — функция пакета, не метод!
func NewCache(client *Client, updateInterval time.Duration) *Cache {
	c := &Cache{
		client: client,
		data:   make(map[string]interface{}),
		tags:   make(map[string]struct{}),
	}
	go c.runUpdater(updateInterval)
	return c
}

// RegisterTag добавляет тег в список для обновления
func (c *Cache) RegisterTag(tag string) {
	c.mu.Lock()
	c.tags[tag] = struct{}{}
	c.mu.Unlock()
}

// Get читает из кэша (быстро!)
func (c *Cache) Get(tag string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[tag]
	return val, ok
}

// Set пишет в кэш и в PLC (асинхронно)
func (c *Cache) Set(tag string, value interface{}) {
	// Пишем в PLC (блокирующий, но быстрый)
	c.client.Write(tag, value)

	// Обновляем кэш
	c.mu.Lock()
	c.data[tag] = value
	c.mu.Unlock()
}

// GetBit читает бит из кэшированного DINT/INT
func (c *Cache) GetBit(tag string, bit int) (bool, error) {
	val, ok := c.Get(tag)
	if !ok {
		return false, nil // тег не в кэше, вернём false
	}

	var intVal int32
	switch v := val.(type) {
	case int16:
		intVal = int32(v)
	case int32:
		intVal = v
	case uint16:
		intVal = int32(v)
	case uint32:
		intVal = int32(v)
	default:
		return false, nil
	}

	return ((intVal >> bit) & 1) == 1, nil
}

// SetBit модифицирует бит и пишет в PLC
func (c *Cache) SetBit(tag string, bit int, value bool) error {
	val, ok := c.Get(tag)
	if !ok {
		// Читаем напрямую если нет в кэше
		var raw any
		if err := c.client.Read(tag, &raw); err != nil {
			return err
		}
		val = raw
	}

	var intVal int32
	switch v := val.(type) {
	case int16:
		intVal = int32(v)
	case int32:
		intVal = v
	case uint16:
		intVal = int32(v)
	case uint32:
		intVal = int32(v)
	default:
		return nil
	}

	if value {
		intVal |= (1 << bit)
	} else {
		intVal &^= (1 << bit)
	}

	c.Set(tag, intVal)
	return nil
}

// Фоновое обновление всех зарегистрированных тегов
func (c *Cache) runUpdater(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.RLock()
		tags := make([]string, 0, len(c.tags))
		for tag := range c.tags {
			tags = append(tags, tag)
		}
		c.mu.RUnlock()

		// Читаем все теги батчем (или по очереди)
		for _, tag := range tags {
			var val any
			if err := c.client.Read(tag, &val); err == nil {
				c.mu.Lock()
				c.data[tag] = val
				c.mu.Unlock()
			}
		}
	}
}

// GetAll возвращает копию всех данных (для web API)
func (c *Cache) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// GetStats возвращает статистику
func (c *Cache) GetStats() (tagCount int, lastUpdate time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.tags), time.Now() // TODO: хранить реальное время последнего обновления
}
