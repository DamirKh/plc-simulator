-- scripts/scada_async.lua
-- Асинхронные команды SCADA (аналог QThread)

local M = {}

-- Хранилище команд по имени
local commands = {}
-- Список для итерации (порядок не важен)
local command_list = {}

-- Инициализация из конфигурации
function M.init(config_scada)
    for name, cfg in pairs(config_scada.commands) do
        -- Регистрируем тег в кэше PLC
        local err = plc_register_tag(cfg.tag)
        if err then
            log("ERROR registering " .. name .. ": " .. err)
        else
            log("Registered SCADA: " .. name .. " -> " .. cfg.tag)
        end
        
        -- Создаём команду
        commands[name] = {
            bit = cfg,
            name = name,
            pulse_ms = cfg.pulse_ms or 3000,
            stop_time = 0,
            running = false,
        }
        table.insert(command_list, commands[name])
    end
    
    -- Регистрируем статусы тоже
    for name, st in pairs(config_scada.statuses or {}) do
        local err = plc_register_tag(st.tag)
        if err then
            log("ERROR registering status " .. name .. ": " .. err)
        else
            log("Registered SCADA status: " .. name .. " -> " .. st.tag)
        end
    end
    
    log("SCADA: loaded " .. #command_list .. " commands")
end

-- Получить команду по имени
function M.get(name)
    return commands[name]
end

-- Запустить команду (аналог __call__)
function M.raise(cmd_or_name)
    local cmd
    if type(cmd_or_name) == "string" then
        cmd = commands[cmd_or_name]
        if not cmd then
            log("[SCADA ERROR] Unknown command: " .. cmd_or_name)
            return false
        end
    else
        cmd = cmd_or_name
    end
    
    cmd.stop_time = os.clock() * 1000 + cmd.pulse_ms
    cmd.running = true
    log(string.format("[SCADA] %s raised for %d ms", cmd.name, cmd.pulse_ms))
    return true
end

-- Обновить все команды (вызывать каждый цикл из main)
function M.update_all(dt_ms)
    for _, cmd in ipairs(command_list) do
        if cmd.running then
            local now = os.clock() * 1000
            
            if now < cmd.stop_time then
                -- Импульс активен
                local err = plc_write_bit(cmd.bit.tag, cmd.bit.bit, true)
                if err then
                    log("[SCADA ERROR] " .. err)
                end
            else
                -- Время вышло, сбрасываем
                local err = plc_write_bit(cmd.bit.tag, cmd.bit.bit, false)
                if not err then
                    cmd.running = false
                    log(string.format("[SCADA] %s dropped", cmd.name))
                end
            end
        end
    end
end

return M
