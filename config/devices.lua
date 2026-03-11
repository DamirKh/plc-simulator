-- Конфигурация полевого оборудования и SCADA точек
-- Плоская структура без функций-помощников

return {
    devices = {
        xv1301 = {
            name = "XV1301",
            type = "motorized_valve",
            params = {
                open_time_ms = 15000,
                close_time_ms = 15000,
            },
            io = {
                open_cmd   = {tag = "RCH04:1:O.Data", bit = 19, inverted = false, type = "output" },
                close_cmd  = {tag = "RCH04:1:O.Data", bit = 18, type = "output"},
                stop_cmd   = {tag = "RCH04:1:O.Data", bit = 20, type = "output"},
                opened_fb  = {tag = "RCH03:2:I.Data", bit = 7,  type = "input"},
                closed_fb  = {tag = "RCH03:2:I.Data", bit = 6,  type = "input"},
                ready      = {tag = "RCH03:2:I.Data", bit = 11, type = "input"},
            },
        },

        xv1302 = {
            name = "XV1302",
            type = "motorized_valve",
            params = {
                open_time_ms = 15000,
                close_time_ms = 15000,
            },
            io = {
                open_cmd   = {tag = "RCH04:1:O.Data", bit = 22, type = "output"},
                close_cmd  = {tag = "RCH04:1:O.Data", bit = 21, type = "output"},
                stop_cmd   = {tag = "RCH04:1:O.Data", bit = 23, type = "output"},
                opened_fb  = {tag = "RCH03:2:I.Data", bit = 19,  type = "input"},
                closed_fb  = {tag = "RCH03:2:I.Data", bit = 18,  type = "input"},
                ready      = {tag = "RCH03:2:I.Data", bit = 23, type = "input"},
            },
        },
    },

    scada = {
        commands = {
            xv1301_reset = {tag = "N68[227]", bit = 0, type = "input"},
            xv1302_reset = {tag = "N68[227]", bit = 1, type = "input"},
        },

        statuses = {
            system_ready = {tag = "N68[228]", bit = 0, type = "output"},
        },
    },
}
