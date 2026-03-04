package field

import (
	"fmt"
	"plc-simulator/pkg/io"
	"time"
)

type ValveState int

const (
	ValveClosed ValveState = iota
	ValveOpening
	ValveOpened
	ValveClosing
	ValveStopped
)

// MotorizedValve симулирует моторизованную задвижку
type MotorizedValve struct {
	Name string

	// IO с точки зрения PLC:
	// Команды от PLC к полю (мы ЧИТАЕМ)
	OpenCmd  io.DiscreteReader
	CloseCmd io.DiscreteReader
	StopCmd  io.DiscreteReader

	// Фидбек от поля к PLC (мы ПИШЕМ)
	OpenedFB io.DiscreteWriter
	ClosedFB io.DiscreteWriter
	Ready    io.DiscreteWriter

	// Параметры
	OpenTime  time.Duration
	CloseTime time.Duration

	// Состояние
	State    ValveState
	Position float64 // 0.0 - 100.0%
}

func NewMotorizedValve(name string, openTime, closeTime time.Duration) *MotorizedValve {
	return &MotorizedValve{
		Name:      name,
		OpenTime:  openTime,
		CloseTime: closeTime,
		State:     ValveClosed,
		Position:  0.0,
	}
}

// Update выполняет один цикл симуляции
func (v *MotorizedValve) Update(dt time.Duration) error {
	// Читаем команды от PLC
	openCmd, _ := v.OpenCmd.Read()
	closeCmd, _ := v.CloseCmd.Read()
	stopCmd, _ := v.StopCmd.Read()

	// State machine
	switch v.State {
	case ValveClosed:
		if openCmd && !closeCmd {
			v.State = ValveOpening
		}

	case ValveOpening:
		if stopCmd {
			v.State = ValveStopped
			break
		}
		if closeCmd {
			v.State = ValveClosing
			break
		}

		v.Position += 100.0 * dt.Seconds() / v.OpenTime.Seconds()
		if v.Position >= 100.0 {
			v.Position = 100.0
			v.State = ValveOpened
		}

	case ValveOpened:
		if closeCmd && !openCmd {
			v.State = ValveClosing
		}

	case ValveClosing:
		if stopCmd {
			v.State = ValveStopped
			break
		}
		if openCmd {
			v.State = ValveOpening
			break
		}

		v.Position -= 100.0 * dt.Seconds() / v.CloseTime.Seconds()
		if v.Position <= 0.0 {
			v.Position = 0.0
			v.State = ValveClosed
		}

	case ValveStopped:
		if openCmd && !closeCmd && v.Position < 100.0 {
			v.State = ValveOpening
		} else if closeCmd && !openCmd && v.Position > 0.0 {
			v.State = ValveClosing
		}
		if v.Position >= 100.0 {
			v.State = ValveOpened
		} else if v.Position <= 0.0 {
			v.State = ValveClosed
		}
	}

	// Пишем фидбек в PLC
	opened := v.State == ValveOpened || (v.State == ValveOpening && v.Position > 95)
	closed := v.State == ValveClosed || (v.State == ValveClosing && v.Position < 5)

	if err := v.OpenedFB.Write(opened); err != nil {
		return fmt.Errorf("%s OpenedFB: %w", v.Name, err)
	}
	if err := v.ClosedFB.Write(closed); err != nil {
		return fmt.Errorf("%s ClosedFB: %w", v.Name, err)
	}
	if err := v.Ready.Write(true); err != nil {
		return fmt.Errorf("%s Ready: %w", v.Name, err)
	}

	return nil
}

func (v *MotorizedValve) String() string {
	stateNames := []string{"Closed", "Opening", "Opened", "Closing", "Stopped"}
	return fmt.Sprintf("%s:%s/%.0f%%", v.Name, stateNames[v.State], v.Position)
}
