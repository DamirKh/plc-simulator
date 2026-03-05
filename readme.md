
## структура проекта

<pre>
plc-simulator/
├── config/
│   └── devices.lua           # Конфигурация устройств на Lua
├── scripts/
│   ├── main.lua              # Главный скрипт
│   ├── valve_sim.lua         # Логика задвижки
│   └── scada_sim.lua         # SCADA симулятор (опционально)
├── pkg/
│   ├── luaapi/
│   │   └── api.go            # Go-функции для Lua (только PLC драйвер)
│   └── plc/
│       └── client.go
└── main.go
</pre>
