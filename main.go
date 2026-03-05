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
	"plc-simulator/pkg/field"
	"plc-simulator/pkg/io"
	"plc-simulator/pkg/plc"
	"plc-simulator/pkg/scada"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	fmt.Printf("Режим: %s, PLC: %s\n", cfg.Mode, cfg.PLC.Path)

	client := plc.NewClient(cfg.PLC.Path)
	if err := client.Connect(); err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer client.Disconnect()
	fmt.Println("✓ Подключено к PLC")

	// Field Simulator (всегда)
	fieldReg := field.NewRegistry()
	for _, devCfg := range cfg.FieldDevices {
		valve := createValveFromConfig(client, devCfg)
		fieldReg.Add(valve)
		fmt.Printf("  + Field: %s\n", devCfg.Name)
	}

	// SCADA Simulator (если не field_only)
	var scadaSim *scada.Simulator
	if cfg.Mode != config.ModeOff {
		scadaSim = scada.NewSimulator(scada.Mode(cfg.Mode))
		for name, ioCfg := range cfg.SCADA.Commands {
			scadaSim.Commands[name] = inputFromConfig(client, ioCfg)
		}
		fmt.Printf("  + SCADA: %d команд\n", len(scadaSim.Commands))
	}

	fmt.Println("Запуск. Ctrl+C для выхода")

	// Тест: импульс reset для XV1301 (только в simulation)
	if cfg.Mode == config.ModeSimulation {
		go func() {
			time.Sleep(2 * time.Second)
			fmt.Println("\n[TEST] Импульс xv1301_scada_reset...")
			if err := scadaSim.PulseCommand("xv1301_scada_reset", 3*time.Second); err != nil {
				log.Printf("Ошибка: %v", err)
			} else {
				fmt.Println("[TEST] Импульс завершён")
			}
			fmt.Println("\n[TEST] Импульс xv1302_scada_reset...")
			if err := scadaSim.PulseCommand("xv1302_scada_reset", 3*time.Second); err != nil {
				log.Printf("Ошибка: %v", err)
			} else {
				fmt.Println("[TEST] Импульс завершён")
			}
		}()
	}

	// Главный цикл
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nЗавершение...")
		cancel()
	}()

	ticker := time.NewTicker(cfg.CycleTime())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := fieldReg.Update(cfg.CycleTime()); err != nil {
				log.Printf("Field error: %v", err)
			}
			fmt.Printf("\r%s", fieldReg.Status())
		}
	}
}

func createValveFromConfig(client *plc.Client, cfg config.DeviceConfig) *field.MotorizedValve {
	valve := field.NewMotorizedValve(
		cfg.Name,
		time.Duration(cfg.OpenTimeSec)*time.Second,
		time.Duration(cfg.CloseTimeSec)*time.Second,
	)
	valve.OpenCmd = outputFromConfig(client, cfg.IO.OpenCmd)
	valve.CloseCmd = outputFromConfig(client, cfg.IO.CloseCmd)
	valve.StopCmd = outputFromConfig(client, cfg.IO.StopCmd)
	valve.OpenedFB = inputFromConfig(client, cfg.IO.OpenedFB)
	valve.ClosedFB = inputFromConfig(client, cfg.IO.ClosedFB)
	valve.Ready = inputFromConfig(client, cfg.IO.Ready)
	return valve
}

// Для полевого оборудования: output = PLC пишет, мы читаем
func outputFromConfig(client *plc.Client, cfg config.IOBit) io.DiscreteReader {
	if cfg.Type != "output" {
		log.Fatalf("Ожидался type=output, got %s", cfg.Type)
	}
	return io.NewDiscreteOutput(client, cfg.Tag, cfg.Bit, cfg.Inverted)
}

// Для полевого оборудования: input = мы пишем, PLC читает
func inputFromConfig(client *plc.Client, cfg config.IOBit) io.DiscreteWriter {
	if cfg.Type != "input" {
		log.Fatalf("Ожидался type=input, got %s", cfg.Type)
	}
	return io.NewDiscreteInput(client, cfg.Tag, cfg.Bit, cfg.Inverted)
}

// Булевы: читаем из PLC (выходы PLC)
func boolOutputFromConfig(client *plc.Client, cfg config.IOBit) io.BoolReader {
	if cfg.Type != "bool_output" {
		log.Fatalf("Ожидался type=bool_output, got %s", cfg.Type)
	}
	return io.NewBoolReader(client, cfg.Tag)
}

// Булевы: пишем в PLC (входы PLC)
func boolInputFromConfig(client *plc.Client, cfg config.IOBit) io.BoolWriter {
	if cfg.Type != "bool_input" {
		log.Fatalf("Ожидался type=bool_input, got %s", cfg.Type)
	}
	return io.NewBoolWriter(client, cfg.Tag)
}
