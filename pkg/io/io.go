package io

import (
	"fmt"
	"plc-simulator/pkg/plc"
)

// === Интерфейсы ===

type DiscreteReader interface {
	Read() (bool, error)
}

type DiscreteWriter interface {
	Write(value bool) error
}

type BoolReader interface {
	Read() (bool, error)
}

type BoolWriter interface {
	Write(value bool) error
}

// === Реализации ===

// DiscreteOutput читает бит из DINT/INT (PLC пишет, мы читаем)
type DiscreteOutput struct {
	client   *plc.Client
	tag      string
	bit      int
	inverted bool
}

func NewDiscreteOutput(client *plc.Client, tag string, bit int, inverted bool) DiscreteReader {
	return &DiscreteOutput{
		client:   client,
		tag:      tag,
		bit:      bit,
		inverted: inverted,
	}
}

func (do *DiscreteOutput) Read() (bool, error) {
	var raw any
	if err := do.client.Read(do.tag, &raw); err != nil {
		return false, fmt.Errorf("read %s: %w", do.tag, err)
	}

	var val int32
	switch v := raw.(type) {
	case int16:
		val = int32(v)
	case int32:
		val = v
	case uint16:
		val = int32(v)
	case uint32:
		val = int32(v)
	default:
		return false, fmt.Errorf("unsupported type %T for %s", raw, do.tag)
	}

	bitVal := plc.GetBit(val, do.bit)
	if do.inverted {
		return !bitVal, nil
	}
	return bitVal, nil
}

// DiscreteInput пишет бит в DINT/INT (мы пишем, PLC читает)
type DiscreteInput struct {
	client   *plc.Client
	tag      string
	bit      int
	inverted bool
}

func NewDiscreteInput(client *plc.Client, tag string, bit int, inverted bool) DiscreteWriter {
	return &DiscreteInput{
		client:   client,
		tag:      tag,
		bit:      bit,
		inverted: inverted,
	}
}

func (di *DiscreteInput) Write(value bool) error {
	if di.inverted {
		value = !value
	}

	var raw any
	if err := di.client.Read(di.tag, &raw); err != nil {
		return fmt.Errorf("read %s: %w", di.tag, err)
	}

	switch v := raw.(type) {
	case int16:
		return di.client.Write(di.tag, plc.SetBitValue16(v, di.bit, value))
	case int32:
		return di.client.Write(di.tag, plc.SetBitValue(v, di.bit, value))
	case uint16:
		return di.client.Write(di.tag, uint16(plc.SetBitValue16(int16(v), di.bit, value)))
	case uint32:
		return di.client.Write(di.tag, uint32(plc.SetBitValue(int32(v), di.bit, value)))
	default:
		return fmt.Errorf("unsupported type %T for %s", raw, di.tag)
	}
}

// boolReader читает BOOL тег напрямую
type boolReader struct {
	client *plc.Client
	tag    string
}

func NewBoolReader(client *plc.Client, tag string) BoolReader {
	return &boolReader{
		client: client,
		tag:    tag,
	}
}

func (br *boolReader) Read() (bool, error) {
	var val bool
	if err := br.client.Read(br.tag, &val); err != nil {
		return false, fmt.Errorf("read bool %s: %w", br.tag, err)
	}
	return val, nil
}

// boolWriter пишет BOOL тег напрямую
type boolWriter struct {
	client *plc.Client
	tag    string
}

func NewBoolWriter(client *plc.Client, tag string) BoolWriter {
	return &boolWriter{
		client: client,
		tag:    tag,
	}
}

func (bw *boolWriter) Write(value bool) error {
	if err := bw.client.Write(bw.tag, value); err != nil {
		return fmt.Errorf("write bool %s: %w", bw.tag, err)
	}
	return nil
}
