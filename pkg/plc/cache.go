package plc

import (
	"fmt"
	"sync"
	"time"
)

type TagInfo struct {
	Tag      string
	TypeHint string
	LastRead time.Time
}

type Cache struct {
	client     *Client
	data       map[string]interface{}
	mu         sync.RWMutex
	tags       map[string]*TagInfo
	writeQueue map[string]interface{} // очередь на запись
	writeMu    sync.Mutex
}

func NewCache(client *Client, updateInterval time.Duration) *Cache {
	c := &Cache{
		client:     client,
		data:       make(map[string]interface{}),
		tags:       make(map[string]*TagInfo),
		writeQueue: make(map[string]interface{}),
	}
	go c.runUpdater(updateInterval)
	return c
}

// RegisterTag с пробным чтением для определения типа
func (c *Cache) RegisterTag(tag string) error {
	var raw any
	if err := c.client.Read(tag, &raw); err != nil {
		return fmt.Errorf("probe read failed: %w", err)
	}

	typeHint := c.detectType(raw)

	c.mu.Lock()
	c.tags[tag] = &TagInfo{
		Tag:      tag,
		TypeHint: typeHint,
	}
	c.data[tag] = raw
	c.mu.Unlock()

	return nil
}

// GetBit — быстрое чтение из кэша
func (c *Cache) GetBit(tag string, bit int) (bool, error) {
	c.mu.RLock()
	val, ok := c.data[tag]
	c.mu.RUnlock()

	if !ok {
		return false, fmt.Errorf("tag not in cache: %s", tag)
	}

	return c.extractBit(val, bit)
}

// SetBit — проверяет изменение, ставит в очередь на запись
func (c *Cache) SetBit(tag string, bit int, value bool) error {
	// Получаем инфо о теге
	c.mu.RLock()
	info, ok := c.tags[tag]
	if !ok {
		c.mu.RUnlock()
		return fmt.Errorf("tag not registered: %s", tag)
	}
	cachedVal, hasCached := c.data[tag]
	c.mu.RUnlock()

	// Получаем текущее значение (из кэша или читаем)
	var currentVal any
	if hasCached {
		currentVal = cachedVal
	} else {
		currentVal = c.createZero(info.TypeHint)
		if err := c.client.Read(tag, &currentVal); err != nil {
			return err
		}
	}

	// Проверяем, изменился ли бит
	currentBit, _ := c.extractBit(currentVal, bit)
	if currentBit == value {
		// Бит уже в нужном состоянии — не пишем в PLC!
		return nil
	}

	// Модифицируем бит
	newVal, err := c.setBitInValue(currentVal, bit, value)
	if err != nil {
		return err
	}

	// Обновляем кэш сразу (оптимистично)
	c.mu.Lock()
	c.data[tag] = newVal
	c.mu.Unlock()

	// Ставим в очередь на асинхронную запись в PLC
	c.writeMu.Lock()
	c.writeQueue[tag] = newVal
	c.writeMu.Unlock()

	return nil
}

// runUpdater — фоновое чтение и запись
func (c *Cache) runUpdater(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// 1. Обрабатываем очередь записи (в приоритете!)
		c.flushWriteQueue()

		// 2. Читаем все зарегистрированные теги
		c.readAllTags()
	}
}

// flushWriteQueue — пишет накопленные изменения в PLC
func (c *Cache) flushWriteQueue() {
	c.writeMu.Lock()
	queue := make(map[string]interface{}, len(c.writeQueue))
	for k, v := range c.writeQueue {
		queue[k] = v
	}
	c.writeQueue = make(map[string]interface{}) // очистка
	c.writeMu.Unlock()

	for tag, val := range queue {
		if err := c.client.Write(tag, val); err != nil {
			// Возвращаем в очередь на повторную попытку?
			c.writeMu.Lock()
			c.writeQueue[tag] = val
			c.writeMu.Unlock()
		}
	}
}

// readAllTags — батчевое чтение всех тегов
func (c *Cache) readAllTags() {
	c.mu.RLock()
	tags := make([]*TagInfo, 0, len(c.tags))
	for _, info := range c.tags {
		tags = append(tags, info)
	}
	c.mu.RUnlock()

	if len(tags) == 0 {
		return
	}

	// Формируем map для ReadMulti
	readMap := make(map[string]any)
	for _, info := range tags {
		readMap[info.Tag] = c.createZero(info.TypeHint)
	}

	if err := c.client.ReadMulti(readMap); err != nil {
		return // логируем
	}

	// Обновляем кэш (только если нет в очереди на запись — избегаем race)
	c.writeMu.Lock()
	pendingWrites := make(map[string]bool)
	for tag := range c.writeQueue {
		pendingWrites[tag] = true
	}
	c.writeMu.Unlock()

	c.mu.Lock()
	for tag, val := range readMap {
		if !pendingWrites[tag] { // не перезаписываем то, что ещё не записали в PLC
			c.data[tag] = val
		}
	}
	c.mu.Unlock()
}

// GetAll — для web API
func (c *Cache) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// === helper методы ===

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
		return int32(0)
	}
}

func (c *Cache) extractBit(val interface{}, bit int) (bool, error) {
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
		return false, fmt.Errorf("unsupported type %T", val)
	}
	return ((intVal >> bit) & 1) == 1, nil
}

func (c *Cache) setBitInValue(val interface{}, bit int, set bool) (interface{}, error) {
	switch v := val.(type) {
	case int16:
		mask := int16(1 << bit)
		if set {
			return v | mask, nil
		}
		return v &^ mask, nil
	case int32:
		mask := int32(1 << bit)
		if set {
			return v | mask, nil
		}
		return v &^ mask, nil
	case uint16:
		mask := uint16(1 << bit)
		if set {
			return v | mask, nil
		}
		return v &^ mask, nil
	case uint32:
		mask := uint32(1 << bit)
		if set {
			return v | mask, nil
		}
		return v &^ mask, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", val)
	}
}
