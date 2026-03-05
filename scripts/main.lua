local config = require("config.devices")
local valve_sim = require("valve_sim")
local scada_async = require("scada_async")

-- Инициализация задвижек
local valves = {
    xv1301 = valve_sim.new(config.devices.xv1301),
    xv1302 = valve_sim.new(config.devices.xv1302),
}

-- Создаём команды SCADA (аналог Scada_1b_Command)
local cmd_xv1301_reset = scada_async.create_command(
    config.scada.commands.xv1301_reset,
    "xv1301_reset",
    3000  -- 3 сек
)
scada_async.register(cmd_xv1301_reset)

local cmd_xv1302_reset = scada_async.create_command(
    config.scada.commands.xv1302_reset,
    "xv1302_reset",
    3000
)
scada_async.register(cmd_xv1302_reset)

-- Главный цикл (аналог run())
local timer = 0
local dt = 100  -- ms

log("=== Main Loop Started ===")

-- Запускаем первую команду через 2 сек
local test_stage = 0

while true do
    timer = timer + dt
    
    -- === "Поток 1": Обновление задвижек (всегда!) ===
    for name, valve in pairs(valves) do
        valve_sim.update(valve, dt)
    end
    
    -- === "Поток 2": Обновление SCADA команд ===
    scada_async.update_all(dt)
    
    -- === "Поток 3": Тестовый сценарий ===
    if test_stage == 0 and timer >= 2000 then
        scada_async.raise(cmd_xv1301_reset)
        test_stage = 1
    elseif test_stage == 1 and timer >= 6000 then
        scada_async.raise(cmd_xv1302_reset)
        test_stage = 2
    end
    
    -- Вывод статуса
    if (timer % 1000) < dt then
        log(valve_sim.status(valves.xv1301) .. " | " .. valve_sim.status(valves.xv1302))
    end
    
    sleep(dt)
end