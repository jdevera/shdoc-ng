-- LuaSnip snippets for shdoc-ng annotations.
-- Usage: require('shdoc-ng.snippets').setup()
-- Requires LuaSnip to be installed.

local M = {}

function M.setup()
    local ok, ls = pcall(require, 'luasnip')
    if not ok then
        return
    end

    local s = ls.snippet
    local t = ls.text_node
    local i = ls.insert_node

    ls.add_snippets('sh', {
        -- File-level tags
        s('@name', {
            t('@name '),
            i(0, 'library-name'),
        }),
        s('@brief', {
            t('@brief '),
            i(0, 'Short description.'),
        }),
        s('@author', {
            t('@author '),
            i(0, 'Name <email>'),
        }),
        s('@license', {
            t('@license '),
            i(0, 'MIT'),
        }),
        s('@version', {
            t('@version '),
            i(0, '1.0.0'),
        }),
        s('@section', {
            t('@section '),
            i(0, 'Section Name'),
        }),

        -- Function-level tags
        s('@desc', {
            t('@description '),
            i(0),
        }),
        s('@arg', {
            t('@arg $'),
            i(1, '1'),
            t(' '),
            i(2, 'string'),
            t(' '),
            i(3, 'Description.'),
        }),
        s('@opt', {
            t('@option '),
            i(1, '--flag'),
            t(' '),
            i(2, 'Description.'),
        }),
        s('@exit', {
            t('@exitcode '),
            i(1, '0'),
            t(' '),
            i(2, 'Description.'),
        }),
        s('@ex', {
            t({ '@example', '#   ' }),
            i(0),
        }),
        s('@see', {
            t('@see '),
            i(0),
        }),
        s('@warn', {
            t('@warning '),
            i(0),
        }),
        s('@dep', {
            t('@deprecated '),
            i(0),
        }),
        s('@label', {
            t('@label '),
            i(0),
        }),
        s('@noargs', {
            t('@noargs'),
        }),
        s('@internal', {
            t('@internal'),
        }),
        s('@set', {
            t('@set '),
            i(1, 'VAR_NAME'),
            t(' '),
            i(2, 'string'),
            t(' '),
            i(3, 'Description.'),
        }),
        s('@env', {
            t('@env '),
            i(1, 'VAR_NAME'),
            t(' '),
            i(2, 'string'),
            t(' '),
            i(3, 'Description.'),
        }),
        s('@stdin', {
            t('@stdin '),
            i(0, 'Description.'),
        }),
        s('@stdout', {
            t('@stdout '),
            i(0, 'Description.'),
        }),
        s('@stderr', {
            t('@stderr '),
            i(0, 'Description.'),
        }),

        -- Full doc block skeleton (starts a new comment block)
        s('docblock', {
            t({ '# @description ' }),
            i(1, 'Brief description.'),
            t({ '', '# @arg $1 ' }),
            i(2, 'string'),
            t(' '),
            i(3, 'Description.'),
            t({ '', '# @exitcode 0 ' }),
            i(4, 'Success.'),
            t({ '', '# @example', '#   ' }),
            i(5, 'function_name "arg1"'),
            t({ '', '' }),
            i(0),
        }),
    })

    -- Also register for bash and zsh filetypes.
    ls.filetype_extend('bash', { 'sh' })
    ls.filetype_extend('zsh', { 'sh' })
end

return M
