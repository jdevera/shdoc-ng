-- Highlight shdoc-ng annotations using extmarks (works with tree-sitter).

local M = {}

local ns = vim.api.nvim_create_namespace('shdoc-ng')

-- All known tags (including shorthands).
local tags = {
    'description', 'desc', 'arg', 'option', 'opt', 'example', 'exitcode',
    'exit', 'see', 'warning', 'warn', 'deprecated', 'noargs', 'internal',
    'name', 'brief', 'file', 'author', 'license', 'version', 'section',
    'set', 'env', 'stdin', 'stdout', 'stderr',
}

local tag_set = {}
for _, t in ipairs(tags) do
    tag_set[t] = true
end

-- Pattern: captures the @ and tag name from a comment line.
local tag_pattern = '^(%s*#%s*)@(%w+)'

-- Apply highlights to a single buffer.
function M.apply(buf)
    buf = buf or 0
    vim.api.nvim_buf_clear_namespace(buf, ns, 0, -1)

    local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
    for i, line in ipairs(lines) do
        local prefix, tag = line:match(tag_pattern)
        if tag and tag_set[tag] then
            local at_col = #prefix
            local tag_end = at_col + 1 + #tag

            -- Highlight @tag keyword.
            vim.api.nvim_buf_set_extmark(buf, ns, i - 1, at_col, {
                end_col = tag_end,
                hl_group = 'shdocTag',
                priority = 200,
            })

            -- Tag-specific sub-highlights.
            if tag == 'arg' then
                -- Highlight $N or $@ parameter, then type.
                local param, typ = line:match('@arg%s+(%$[0-9@]+)%s+(%S+)', at_col + 1)
                if param then
                    local ps = line:find('%$', tag_end)
                    if ps then
                        vim.api.nvim_buf_set_extmark(buf, ns, i - 1, ps - 1, {
                            end_col = ps - 1 + #param,
                            hl_group = 'shdocArg',
                            priority = 200,
                        })
                        -- Type after param.
                        if typ then
                            local ts = line:find(typ, ps + #param, true)
                            if ts then
                                vim.api.nvim_buf_set_extmark(buf, ns, i - 1, ts - 1, {
                                    end_col = ts - 1 + #typ,
                                    hl_group = 'shdocType',
                                    priority = 200,
                                })
                            end
                        end
                    end
                end
            elseif tag == 'option' or tag == 'opt' then
                -- Highlight flags: -f, --flag (only preceded by space, | or /).
                local pos = tag_end + 1
                while pos <= #line do
                    local fs, fe = line:find('%-%-?[a-zA-Z][a-zA-Z0-9_%-]*', pos)
                    if not fs then break end
                    local before = fs > 1 and line:sub(fs - 1, fs - 1) or ''
                    if before == ' ' or before == '|' or before == '/' or before == '\t' then
                        vim.api.nvim_buf_set_extmark(buf, ns, i - 1, fs - 1, {
                            end_col = fe,
                            hl_group = 'shdocFlag',
                            priority = 200,
                        })
                    end
                    pos = fe + 1
                end
                -- Highlight <placeholder> arguments.
                pos = tag_end + 1
                while pos <= #line do
                    local ps, pe = line:find('<[^>]+>', pos)
                    if not ps then break end
                    vim.api.nvim_buf_set_extmark(buf, ns, i - 1, ps - 1, {
                        end_col = pe,
                        hl_group = 'shdocVar',
                        priority = 200,
                    })
                    pos = pe + 1
                end
            elseif tag == 'exitcode' or tag == 'exit' then
                -- Highlight exit code number.
                local num = line:match('@%w+%s+([>!]?%d+)', at_col + 1)
                if num then
                    local ns_pos = line:find(num, tag_end, true)
                    if ns_pos then
                        vim.api.nvim_buf_set_extmark(buf, ns, i - 1, ns_pos - 1, {
                            end_col = ns_pos - 1 + #num,
                            hl_group = 'shdocNumber',
                            priority = 200,
                        })
                    end
                end
            elseif tag == 'set' or tag == 'env' then
                -- Highlight variable name and optional type.
                local var, typ = line:match('@%w+%s+(%S+)%s+(%S+)', at_col + 1)
                if not var then
                    var = line:match('@%w+%s+(%S+)', at_col + 1)
                end
                if var then
                    local vs = line:find(var, tag_end, true)
                    if vs then
                        vim.api.nvim_buf_set_extmark(buf, ns, i - 1, vs - 1, {
                            end_col = vs - 1 + #var,
                            hl_group = 'shdocVar',
                            priority = 200,
                        })
                        if typ then
                            local ts = line:find(typ, vs + #var, true)
                            if ts then
                                vim.api.nvim_buf_set_extmark(buf, ns, i - 1, ts - 1, {
                                    end_col = ts - 1 + #typ,
                                    hl_group = 'shdocType',
                                    priority = 200,
                                })
                            end
                        end
                    end
                end
            end
        end
    end
end

-- Define default highlight groups.
function M.define_groups()
    vim.api.nvim_set_hl(0, 'shdocTag', { default = true, link = 'Keyword' })
    vim.api.nvim_set_hl(0, 'shdocArg', { default = true, link = 'Identifier' })
    vim.api.nvim_set_hl(0, 'shdocFlag', { default = true, link = 'Function' })
    vim.api.nvim_set_hl(0, 'shdocNumber', { default = true, link = 'Number' })
    vim.api.nvim_set_hl(0, 'shdocVar', { default = true, link = 'Identifier' })
    vim.api.nvim_set_hl(0, 'shdocType', { default = true, link = 'Type' })
end

return M
