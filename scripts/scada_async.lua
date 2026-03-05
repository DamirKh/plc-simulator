-- Асинхронные команды SCADA (аналог QThread)

local M = {}

-- Активные команды (таблица объектов-команд)
local active_commands = {}

-- Создать команду (аналог __init__)
function M.create_command(bit_cfg, name, pulse_time_ms)
    return {
        bit = bit_cfg,
        name = name,
        pulse_time_ms = pulse_time_ms,
        stop_time = 0,
        running = false,
    }
end

-- Запустить команду (аналог __call__)
function M.raise(cmd)
    cmd.stop_time = os.clock() * 1000 + cmd.pulse_time_ms
    cmd.running = true
    log(string.format("[SCADA] Command %s raised for %d ms", cmd.name, cmd.pulse_time_ms))
end

-- Обновить все команды (вызывать каждый цикл из main)
function M.update_all(dt_ms)
    for _, cmd in ipairs(active_commands) do
        if cmd.running then
            local now = os.clock() * 1000
            
            if now < cmd.stop_time then
                -- Импульс активен
                local err = plc_write_bit(cmd.bit.tag, cmd.bit.bit, true)
                if err then
                    log("[SCADA ERROR] " .. err)
                end
            else
                -- Время вышли, сбрасываем
                local err = plc_write_bit(cmd.bit.tag, cmd.bit.bit, false)
                if not err then
                    cmd.running = false
                    log(string.format("[SCADA] Command %s dropped", cmd.name))
                end
            end
        end
    end
end

-- Добавить команду в список активных
function M.register(cmd)
    table.insert(active_commands, cmd)
end

return M
