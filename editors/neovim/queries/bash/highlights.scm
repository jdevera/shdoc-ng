; shdoc-ng tag keywords inside comments: @description, @arg, etc.
; This is a progressive enhancement for users with nvim-treesitter.
; The after/syntax/sh.vim file provides a fallback for all Vim/Neovim users.
((comment) @_c
  (#match? @_c "@(description|desc|arg|option|opt|example|exitcode|exit|see|warning|warn|deprecated|noargs|internal|name|brief|file|author|license|version|section|set|env|stdin|stdout|stderr)")
  (#offset! @_c 0 0 0 0))
