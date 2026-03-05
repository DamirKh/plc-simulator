-- Симуляция моторизованной задвижки

local M = {}

-- Создание задвижки из конфигурации
function M.new(cfg)
    local valve = {
        cfg = cfg,
        position = 0.0,      -- 0-100%
        state = "closed",    -- closed, opening, opened, closing, stopped
    }
    
    log(string.format("Valve %s created: open=%dms, close=%dms",
        cfg.name, cfg.params.open_time_ms, cfg.params.close_time_ms))
    
    return valve
end

-- Чтение команд
function M.read_commands(valve)
    local io = valve.cfg.io
    return {
        open  = plc_read_bit(io.open_cmd.tag, io.open_cmd.bit),
        close = plc_read_bit(io.close_cmd.tag, io.close_cmd.bit),
        stop  = plc_read_bit(io.stop_cmd.tag, io.stop_cmd.bit),
    }
end

-- Запись фидбека
function M.write_feedback(valve)
    local io = valve.cfg.io
    local opened = (valve.state == "opened") or 
                   (valve.state == "opening" and valve.position > 95.0)
    local closed = (valve.state == "closed") or 
                   (valve.state == "closing" and valve.position < 5.0)
    
    plc_write_bit(io.opened_fb.tag, io.opened_fb.bit, opened)
    plc_write_bit(io.closed_fb.tag, io.closed_fb.bit, closed)
    plc_write_bit(io.ready.tag, io.ready.bit, true)
end

-- Обновление задвижки (один цикл)
function M.update(valve, dt_ms)
    local cmds = M.read_commands(valve)
    local p = valve.cfg.params
    
    -- State machine
    if valve.state == "closed" then
        if cmds.open and not cmds.close then
            valve.state = "opening"
            log(valve.cfg.name .. ": OPENING started")
        end
        
    elseif valve.state == "opening" then
        if cmds.stop then
            valve.state = "stopped"
            log(valve.cfg.name .. ": STOPPED at " .. string.format("%.1f", valve.position) .. "%")
        elseif cmds.close then
            valve.state = "closing"
        else
            valve.position = valve.position + (100.0 * dt_ms / p.open_time_ms)
            if valve.position >= 100.0 then
                valve.position = 100.0
                valve.state = "opened"
                log(valve.cfg.name .. ": OPENED")
            end
        end
        
    elseif valve.state == "opened" then
        if cmds.close and not cmds.open then
            valve.state = "closing"
            log(valve.cfg.name .. ": CLOSING started")
        end
        
    elseif valve.state == "closing" then
        if cmds.stop then
            valve.state = "stopped"
        elseif cmds.open then
            valve.state = "opening"
        else
            valve.position = valve.position - (100.0 * dt_ms / p.close_time_ms)
            if valve.position <= 0.0 then
                valve.position = 0.0
                valve.state = "closed"
                log(valve.cfg.name .. ": CLOSED")
            end
        end
        
    elseif valve.state == "stopped" then
        if cmds.open and not cmds.close and valve.position < 100.0 then
            valve.state = "opening"
        elseif cmds.close and not cmds.open and valve.position > 0.0 then
            valve.state = "closing"
        end
    end
    
    -- Пишем фидбек
    M.write_feedback(valve)
    
    return valve.state, valve.position
end

-- Строка состояния
function M.status(valve)
    return string.format("%s:%s/%.0f%%", valve.cfg.name, valve.state, valve.position)
end

return M