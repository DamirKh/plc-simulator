-- SCADA симулятор
print("SCADA simulation loading...")
local M = {}

M.mode = "simulation"  -- simulation | passive

function M.set_mode(mode)
    M.mode = mode
    log("SCADA mode: " .. mode)
end

-- Запись команды (только в simulation)
function M.write_command(cmd_cfg, value)
    if M.mode ~= "simulation" then
        return false, "not in simulation mode"
    end
    
    local err = plc_write_bit(cmd_cfg.tag, cmd_cfg.bit, value)
    if err then
        return false, err
    end
    return true
end

-- Импульс команды
function M.pulse_command(cmd_cfg, duration_ms)
    local ok, err = M.write_command(cmd_cfg, true)
    if not ok then return false, err end
    
    sleep(duration_ms)
    
    ok, err = M.write_command(cmd_cfg, false)
    if not ok then return false, err end
    
    return true
end

-- Чтение статуса (всегда)
function M.read_status(status_cfg)
    local val, err = plc_read_bit(status_cfg.tag, status_cfg.bit)
    if err then
        return false, err
    end
    return val
end

return M
