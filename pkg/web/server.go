// pkg/web/server.go
package web

import (
	"fmt"
	"plc-simulator/pkg/plc"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cache     *plc.Cache
	client    *plc.Client // ← для реального статуса PLC
	startTime time.Time
	port      string
}

func NewServer(cache *plc.Cache, client *plc.Client, port string) *Server {
	return &Server{
		cache:     cache,
		client:    client,
		startTime: time.Now(),
		port:      port,
	}
}

func (s *Server) Run() {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// API
	r.GET("/api/status", s.handleStatus)
	r.GET("/api/cache", s.handleCache)
	r.GET("/api/metrics", s.handleMetrics) // ← новый endpoint
	r.GET("/api/plc", s.handlePLC)
	r.GET("/api/health", s.handleHealth)

	// Static files
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/templates/index.html")

	go func() {
		fmt.Printf("Web server: http://localhost:%s\n", s.port)
		if err := r.Run(":" + s.port); err != nil {
			fmt.Printf("Web error: %v\n", err)
		}
	}()
}

// handleMetrics — метрики производительности кэша
func (s *Server) handleMetrics(c *gin.Context) {
	m := s.cache.GetMetrics()

	// Вычисляем средние значения
	var avgUpdater, avgFlush time.Duration
	if m.UpdaterRuns > 0 {
		avgUpdater = m.UpdaterTotalTime / time.Duration(m.UpdaterRuns)
	}
	if m.FlushRuns > 0 {
		avgFlush = m.FlushTotalTime / time.Duration(m.FlushRuns)
	}

	c.JSON(200, gin.H{
		"updater": gin.H{
			"last_ms":  m.LastUpdaterDuration.Milliseconds(),
			"avg_ms":   avgUpdater.Milliseconds(),
			"total_ms": m.UpdaterTotalTime.Milliseconds(),
			"runs":     m.UpdaterRuns,
		},
		"flush": gin.H{
			"last_ms":  m.LastFlushDuration.Milliseconds(),
			"avg_ms":   avgFlush.Milliseconds(),
			"total_ms": m.FlushTotalTime.Milliseconds(),
			"runs":     m.FlushRuns,
		},
		"write_queue": m.WriteQueueSize,
	})
}

func (s *Server) handleStatus(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(200, gin.H{
		"uptime":     time.Since(s.startTime).Round(time.Second).String(),
		"goroutines": runtime.NumGoroutine(),
		"memory_mb":  m.Alloc / 1024 / 1024,
		"gc_count":   m.NumGC,
	})
}

func (s *Server) handleCache(c *gin.Context) {
	data := s.cache.GetAll()
	c.JSON(200, gin.H{
		"tags_count": len(data),
		"tags":       data,
	})
}

// РЕАЛЬНЫЙ статус PLC
func (s *Server) handlePLC(c *gin.Context) {
	connected := s.client.IsConnected()
	session := "inactive"
	if connected {
		session = "active"
	}

	c.JSON(200, gin.H{
		"connected":     connected,
		"controller_ip": s.client.GetPath(), // ← нужно добавить метод в Client
		"session":       session,
		"last_error":    "",
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
