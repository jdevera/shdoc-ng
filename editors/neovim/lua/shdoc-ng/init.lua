local M = {}
local hl = require('shdoc-ng.highlight')

local function apply(buf, cmd)
    local file = vim.api.nvim_buf_get_name(buf)

    -- Start the LSP server.
    vim.lsp.start({
        name = 'shdoc-lsp',
        cmd = { cmd },
        root_dir = vim.fs.dirname(
            vim.fs.find({ '.git' }, { path = file, upward = true })[1]
        ) or vim.fn.getcwd(),
    }, { bufnr = buf })

    -- Apply annotation highlights.
    hl.apply(buf)
end

function M.setup(opts)
    opts = opts or {}
    local cmd = opts.cmd or 'shdoc-lsp'
    local shell_fts = { sh = true, bash = true, zsh = true }

    hl.define_groups()

    -- Register for future buffers.
    vim.api.nvim_create_autocmd('FileType', {
        pattern = { 'sh', 'bash', 'zsh' },
        callback = function(ev)
            apply(ev.buf, cmd)
        end,
    })

    -- Re-apply highlights on text change.
    vim.api.nvim_create_autocmd({ 'TextChanged', 'TextChangedI' }, {
        pattern = { '*.sh', '*.bash', '*.zsh' },
        callback = function(ev)
            hl.apply(ev.buf)
        end,
    })

    -- Apply to any already-open buffers.
    for _, buf in ipairs(vim.api.nvim_list_bufs()) do
        if vim.api.nvim_buf_is_loaded(buf) and shell_fts[vim.bo[buf].filetype] then
            apply(buf, cmd)
        end
    end
end

return M
