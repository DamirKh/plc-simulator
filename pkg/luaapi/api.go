package luaapi

import (
	"fmt"
	"time"

	"plc-simulator/pkg/plc"

	lua "github.com/yuin/gopher-lua"
)

type State struct {
	PLC *plc.Client
}

func NewState(plc *plc.Client) *State {
	return &State{PLC: plc}
}

// функции, которые будут доступны в Lua
func (s *State) Register(L *lua.LState) {
	// PLC чтение/запись DINT
	L.SetGlobal("plc_read_dint", L.NewFunction(s.plcReadDINT))
	L.SetGlobal("plc_write_dint", L.NewFunction(s.plcWriteDINT))

	// Битовые операции (новое!)
	L.SetGlobal("plc_read_bit", L.NewFunction(s.plcReadBit))
	L.SetGlobal("plc_write_bit", L.NewFunction(s.plcWriteBit))

	// Вспомогательные
	L.SetGlobal("sleep", L.NewFunction(s.sleep))
	L.SetGlobal("log", L.NewFunction(s.log))
}

func (s *State) plcReadDINT(L *lua.LState) int {
	tag := L.CheckString(1)
	var val int32
	if err := s.PLC.Read(tag, &val); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LNumber(val))
	L.Push(lua.LNil)
	return 2
}

func (s *State) plcWriteDINT(L *lua.LState) int {
	tag := L.CheckString(1)
	val := int32(L.CheckNumber(2))
	if err := s.PLC.Write(tag, val); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

// plc_write_bit — автоопределение типа через any
func (s *State) plcWriteBit(L *lua.LState) int {
	tag := L.CheckString(1)
	bit := L.CheckInt(2)
	value := L.CheckBool(3)

	// Читаем с автоопределением типа
	var raw any
	if err := s.PLC.Read(tag, &raw); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	// Модифицируем бит в зависимости от типа
	switch v := raw.(type) {
	case int16:
		var val int16 = v
		if value {
			val |= int16(1 << bit)
		} else {
			val &^= int16(1 << bit)
		}
		if err := s.PLC.Write(tag, val); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

	case int32:
		var val int32 = v
		if value {
			val |= (1 << bit)
		} else {
			val &^= (1 << bit)
		}
		if err := s.PLC.Write(tag, val); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

	case uint16:
		var val uint16 = v
		if value {
			val |= uint16(1 << bit)
		} else {
			val &^= uint16(1 << bit)
		}
		if err := s.PLC.Write(tag, val); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

	case uint32:
		var val uint32 = v
		if value {
			val |= uint32(1 << bit)
		} else {
			val &^= uint32(1 << bit)
		}
		if err := s.PLC.Write(tag, val); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

	default:
		L.Push(lua.LString(fmt.Sprintf("unsupported type %T for tag %s", raw, tag)))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// plc_read_bit — автоопределение типа
func (s *State) plcReadBit(L *lua.LState) int {
	tag := L.CheckString(1)
	bit := L.CheckInt(2)

	var raw any
	if err := s.PLC.Read(tag, &raw); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var result int32

	switch v := raw.(type) {
	case int16:
		result = int32(v)
	case int32:
		result = v
	case uint16:
		result = int32(v)
	case uint32:
		result = int32(v)
	default:
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("unsupported type %T for tag %s", raw, tag)))
		return 2
	}

	bitVal := (result >> bit) & 1
	L.Push(lua.LBool(bitVal == 1))
	L.Push(lua.LNil)
	return 2
}

func (s *State) sleep(L *lua.LState) int {
	ms := L.CheckInt64(1)
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

func (s *State) log(L *lua.LState) int {
	msg := L.CheckString(1)
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), msg)
	return 0
}
