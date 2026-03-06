package luaapi

import (
	"fmt"
	"log"
	"time"

	"plc-simulator/pkg/plc"

	lua "github.com/yuin/gopher-lua"
)

type State struct {
	PLC   *plc.Client
	Cache *plc.Cache
}

func NewState(plcClient *plc.Client) *State {
	cache := plc.NewCache(plcClient, 50*time.Millisecond)
	return &State{
		PLC:   plcClient,
		Cache: cache,
	}
}

func (s *State) Register(L *lua.LState) {
	// Регистрация тегов
	L.SetGlobal("plc_register_tag", L.NewFunction(s.plcRegisterTag))

	// Быстрые операции с кэшем
	L.SetGlobal("plc_read_bit", L.NewFunction(s.plcReadBit))
	L.SetGlobal("plc_write_bit", L.NewFunction(s.plcWriteBit))

	// Для отладки: прямое чтение
	L.SetGlobal("plc_read_dint", L.NewFunction(s.plcReadDINT))

	L.SetGlobal("sleep", L.NewFunction(s.sleep))
	L.SetGlobal("log", L.NewFunction(s.log))
}

func (s *State) plcRegisterTag(L *lua.LState) int {
	tag := L.CheckString(1)

	if err := s.Cache.RegisterTag(tag); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	log.Printf("[CACHE] Registered: %s", tag)
	L.Push(lua.LNil)
	return 1
}

// Быстрое чтение бита из кэша
func (s *State) plcReadBit(L *lua.LState) int {
	tag := L.CheckString(1)
	bit := L.CheckInt(2)

	val, _ := s.Cache.GetBit(tag, bit)
	L.Push(lua.LBool(val))
	L.Push(lua.LNil)
	return 2
}

// Быстрая запись бита через кэш
func (s *State) plcWriteBit(L *lua.LState) int {
	tag := L.CheckString(1)
	bit := L.CheckInt(2)
	value := L.CheckBool(3)

	err := s.Cache.SetBit(tag, bit, value)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

// Прямое чтение DINT (для отладки)
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
