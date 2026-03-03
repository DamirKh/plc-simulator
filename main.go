package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"plc-simulator/pkg/config"
	"plc-simulator/pkg/devices"
	"plc-simulator/pkg/io"
	"plc-simulator/pkg/plc"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	fmt.Printf("Конфигурация: PLC=%s, устройств=%d\n", cfg.PLC.Path, len(cfg.Devices))

	// Подключаемся к PLC
	client := plc.NewClient(cfg.PLC.Path)
	if err := client.Connect(); err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer client.Disconnect()
	fmt.Println("✓ Подключено!")

	// Создаём реестр и заполняем устройства из конфигурации
	registry := devices.NewRegistry()

	for _, devCfg := range cfg.Devices {
		switch devCfg.Type {
		case "motorized_valve":
			valve := createValveFromConfig(client, devCfg)
			registry.Add(valve)
			fmt.Printf("  + %s (%ds/%ds)\n",
				devCfg.Name, devCfg.OpenTimeSec, devCfg.CloseTimeSec)
		default:
			log.Printf("Неизвестный тип устройства: %s", devCfg.Type)
		}
	}

	// Запускаем симуляцию
	fmt.Println("Симуляция запущена. Ctrl+C для выхода")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nЗавершение...")
		cancel()
	}()

	// Горутина симуляции
	go func() {
		if err := registry.Run(ctx, cfg.CycleTime()); err != nil && err != context.Canceled {
			log.Printf("Ошибка: %v", err)
			cancel()
		}
	}()

	// Вывод статуса
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Printf("\r%s", registry.Status())
		}
	}
}

// createValveFromConfig создаёт задвижку из конфигурации
func createValveFromConfig(client *plc.Client, cfg config.DeviceConfig) *devices.MotorizedValve {
	valve := devices.NewMotorizedValve(
		cfg.Name,
		time.Duration(cfg.OpenTimeSec)*time.Second,
		time.Duration(cfg.CloseTimeSec)*time.Second,
	)

	// Привязываем IO из конфигурации
	valve.OpenCmd = outputFromConfig(client, cfg.IO.OpenCmd)   // DO: читаем
	valve.CloseCmd = outputFromConfig(client, cfg.IO.CloseCmd) // DO: читаем
	valve.StopCmd = outputFromConfig(client, cfg.IO.StopCmd)   // DO: читаем
	valve.OpenedFB = inputFromConfig(client, cfg.IO.OpenedFB)  // DI: пишем
	valve.ClosedFB = inputFromConfig(client, cfg.IO.ClosedFB)  // DI: пишем
	valve.Ready = inputFromConfig(client, cfg.IO.Ready)        // DI: пишем

	return valve
}

// outputFromConfig создаёт DiscreteReader (читаем команду от PLC)
func outputFromConfig(client *plc.Client, cfg config.IOBit) io.DiscreteReader {
	if cfg.Type != "output" {
		log.Fatalf("Ожидался type=output для %s, got %s", cfg.Tag, cfg.Type)
	}
	return io.NewDiscreteOutput(client, cfg.Tag, cfg.Bit, cfg.Inverted)
}

// inputFromConfig создаёт DiscreteWriter (пишем фидбек в PLC)
func inputFromConfig(client *plc.Client, cfg config.IOBit) io.DiscreteWriter {
	if cfg.Type != "input" {
		log.Fatalf("Ожидался type=input для %s, got %s", cfg.Tag, cfg.Type)
	}
	return io.NewDiscreteInput(client, cfg.Tag, cfg.Bit, cfg.Inverted)
}
