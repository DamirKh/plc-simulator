package io

import (
	"fmt"
	"plc-simulator/pkg/plc"
)

// DiscreteOutput — дискретный выход PLC (команда от PLC к устройству)
// Мы ЧИТАЕМ этот сигнал
type DiscreteOutput struct {
	client   *plc.Client
	tag      string
	bit      int
	inverted bool
}

func NewDiscreteOutput(client *plc.Client, tag string, bit int, inverted bool) *DiscreteOutput {
	return &DiscreteOutput{
		client:   client,
		tag:      tag,
		bit:      bit,
		inverted: inverted,
	}
}

// Read читает команду от PLC
func (do *DiscreteOutput) Read() (bool, error) {
	var raw int32
	if err := do.client.Read(do.tag, &raw); err != nil {
		return false, fmt.Errorf("read %s: %w", do.tag, err)
	}
	
	val := plc.GetBit(raw, do.bit)
	if do.inverted {
		return !val, nil
	}
	return val, nil
}

// DiscreteInput — дискретный вход PLC (фидбек от устройства к PLC)
// Мы ПИШЕМ этот сигнал
type DiscreteInput struct {
	client   *plc.Client
	tag      string
	bit      int
	inverted bool
}

func NewDiscreteInput(client *plc.Client, tag string, bit int, inverted bool) *DiscreteInput {
	return &DiscreteInput{
		client:   client,
		tag:      tag,
		bit:      bit,
		inverted: inverted,
	}
}

// Write пишет фидбек в PLC
func (di *DiscreteInput) Write(value bool) error {
	if di.inverted {
		value = !value
	}
	
	// Читаем текущее значение
	var raw int32
	if err := di.client.Read(di.tag, &raw); err != nil {
		return fmt.Errorf("read before write %s: %w", di.tag, err)
	}
	
	// Модифицируем бит
	newVal := plc.SetBitValue(raw, di.bit, value)
	
	// Пишем обратно
	if err := di.client.Write(di.tag, newVal); err != nil {
		return fmt.Errorf("write %s: %w", di.tag, err)
	}
	
	return nil
}
