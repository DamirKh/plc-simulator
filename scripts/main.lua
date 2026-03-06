-- scripts/main.lua
local config = require("config.devices")
local valve_sim = require("valve_sim")
local scada_async = require("scada_async")

-- Инициализация задвижек
local valves = {
    xv1301 = valve_sim.new(config.devices.xv1301),
    xv1302 = valve_sim.new(config.devices.xv1302),
}

-- Регистрируем теги задвижек
for name, valve in pairs(valves) do
    valve_sim.register_tags(valve)
end

-- Инициализация SCADA (автоматически создаёт команды и регистрирует теги)
scada_async.init(config.scada)

-- Главный цикл
local timer = 0
local dt = 100
local test_stage = 0

log("=== Main Loop Started ===")

while true do
    local startTime = os.clock()
    timer = timer + dt
    
    -- Поток 1: Задвижки
    for name, valve in pairs(valves) do
        valve_sim.update(valve, dt)
    end
    
    -- Поток 2: SCADA команды
    scada_async.update_all(dt)
    
    -- Поток 3: Тестовый сценарий
    if test_stage == 0 and timer >= 2000 then
        scada_async.raise("xv1301_reset")  -- по имени!
        test_stage = 1
    elseif test_stage == 1 and timer >= 6000 then
        scada_async.raise("xv1302_reset")
        test_stage = 2
    end
    
    -- Вывод статуса
    if (timer % 1000) < dt then
        log(valve_sim.status(valves.xv1301) .. " | " .. valve_sim.status(valves.xv1302))
    end
    
    sleep(dt)
end
