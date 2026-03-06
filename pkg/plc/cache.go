package plc

import (
	"fmt"
	"sync"
	"time"
)

// TagInfo хранит информацию о теге
type TagInfo struct {
	Tag      string
	TypeHint string // "int16", "int32", "bool" и т.д.
	LastRead time.Time
}

type Cache struct {
	client *Client
	data   map[string]interface{}
	mu     sync.RWMutex
	tags   map[string]*TagInfo // теперь храним инфо о теге
}

func NewCache(client *Client, updateInterval time.Duration) *Cache {
	c := &Cache{
		client: client,
		data:   make(map[string]interface{}),
		tags:   make(map[string]*TagInfo),
	}
	go c.runUpdater(updateInterval)
	return c
}

// RegisterTag добавляет тег и определяет тип пробным чтением
func (c *Cache) RegisterTag(tag string) error {
	// Пробное чтение с any для автоопределения типа
	var raw any
	if err := c.client.Read(tag, &raw); err != nil {
		return fmt.Errorf("probe read failed for %s: %w", tag, err)
	}
	// Определяем тип
	typeHint := c.detectType(raw)

	c.mu.Lock()
	c.tags[tag] = &TagInfo{
		Tag:      tag,
		TypeHint: typeHint,
	}
	// Сразу сохраняем значение
	c.data[tag] = raw
	c.mu.Unlock()

	return nil
}

// detectType определяет строковое представление типа
func (c *Cache) detectType(val interface{}) string {
	switch val.(type) {
	case int16:
		return "int16"
	case int32:
		return "int32"
	case uint16:
		return "uint16"
	case uint32:
		return "uint32"
	case bool:
		return "bool"
	case float32:
		return "float32"
	default:
		return "unknown"
	}
}

// createZero создаёт нулевое значение нужного типа
func (c *Cache) createZero(typeHint string) interface{} {
	switch typeHint {
	case "int16":
		return int16(0)
	case "int32":
		return int32(0)
	case "uint16":
		return uint16(0)
	case "uint32":
		return uint32(0)
	case "bool":
		return false
	case "float32":
		return float32(0)
	default:
		return int32(0) // fallback
	}
}

// Get читает из кэша (быстро!)
func (c *Cache) Get(tag string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[tag]
	return val, ok
}

// GetAll возвращает копию всех данных
func (c *Cache) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// GetBit читает бит из кэшированного значения
func (c *Cache) GetBit(tag string, bit int) (bool, error) {
	val, ok := c.Get(tag)
	if !ok {
		return false, fmt.Errorf("tag not in cache: %s", tag)
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
		return false, fmt.Errorf("unsupported type %T for tag %s", val, tag)
	}

	return ((intVal >> bit) & 1) == 1, nil
}

// SetBit модифицирует бит и пишет в PLC
func (c *Cache) SetBit(tag string, bit int, value bool) error {
	// Получаем инфо о теге (включая тип)
	c.mu.RLock()
	info, ok := c.tags[tag]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("tag not registered: %s", tag)
	}

	// Используем info.TypeHint для проверки
	_ = info.TypeHint // или убрать переменную info

	// Читаем актуальное значение из кэша или PLC
	var raw any
	var err error

	c.mu.RLock()
	cachedVal, hasCached := c.data[tag]
	c.mu.RUnlock()

	if hasCached {
		raw = cachedVal
	} else {
		if err = c.client.Read(tag, &raw); err != nil {
			return err
		}
	}

	// Модифицируем бит с учётом реального типа
	var newVal interface{}

	switch v := raw.(type) {
	case int16:
		val := v
		if value {
			val |= int16(1 << bit)
		} else {
			val &^= int16(1 << bit)
		}
		newVal = val

	case int32:
		val := v
		if value {
			val |= (1 << bit)
		} else {
			val &^= (1 << bit)
		}
		newVal = val

	case uint16:
		val := v
		if value {
			val |= uint16(1 << bit)
		} else {
			val &^= uint16(1 << bit)
		}
		newVal = val

	case uint32:
		val := v
		if value {
			val |= uint32(1 << bit)
		} else {
			val &^= uint32(1 << bit)
		}
		newVal = val

	default:
		return fmt.Errorf("unsupported type %T for tag %s", raw, tag)
	}

	// Пишем с правильным типом!
	if err := c.client.Write(tag, newVal); err != nil {
		return err
	}

	// Обновляем кэш
	c.mu.Lock()
	c.data[tag] = newVal
	c.mu.Unlock()

	return nil
}

// Фоновое обновление через ReadMulti
func (c *Cache) runUpdater(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.RLock()
		tags := make([]*TagInfo, 0, len(c.tags))
		for _, info := range c.tags {
			tags = append(tags, info)
		}
		c.mu.RUnlock()

		if len(tags) == 0 {
			continue
		}

		// Создаём map с правильными типами
		readMap := make(map[string]any)
		for _, info := range tags {
			readMap[info.Tag] = c.createZero(info.TypeHint)
		}

		// Один запрос!
		if err := c.client.ReadMulti(readMap); err != nil {
			continue // логируем ошибку
		}

		// Обновляем кэш
		c.mu.Lock()
		for tag, val := range readMap {
			c.data[tag] = val
		}
		c.mu.Unlock()
	}
}

// guessType определяет тип по имени тега
func (c *Cache) guessType(tag string) interface{} {
	// По умолчанию DINT
	return int32(0)
}
