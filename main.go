package main

import (
	"log"

	"plc-simulator/pkg/luaapi"
	"plc-simulator/pkg/plc"

	lua "github.com/yuin/gopher-lua"
)

func main() {
	// Подключаемся к PLC
	client := plc.NewClient("192.168.0.40")
	if err := client.Connect(); err != nil {
		log.Fatal("PLC connect error:", err)
	}
	defer client.Disconnect()

	// Lua VM
	L := lua.NewState()
	defer L.Close()

	// Добавляем пути для require (важно!)
	L.DoString(`
		package.path = package.path .. ";./scripts/?.lua;./config/?.lua;./?.lua"
	`)

	// Регистрируем Go-функции
	api := luaapi.NewState(client)
	api.Register(L)

	// Запускаем main.lua
	if err := L.DoFile("scripts/main.lua"); err != nil {
		log.Fatal("Lua error:", err)
	}
}
