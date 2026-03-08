" Syntax highlighting for shdoc-ng annotations inside shell comments.
" Works in both Vim and Neovim, independent of Tree-sitter.

" Tag keywords (including shorthands)
syntax match shdocTag /@\(description\|desc\|arg\|option\|opt\|example\|exitcode\|exit\|see\|warning\|warn\|deprecated\|noargs\|internal\|name\|brief\|file\|author\|license\|version\|section\|set\|env\|stdin\|stdout\|stderr\)\b/ contained containedin=shComment

" Positional parameters — only after @arg
syntax match shdocArg /\$[0-9]\+\|\$@/ contained containedin=shComment

" Option flags — only on @option/@opt lines
syntax match shdocFlag /--\?[a-zA-Z][a-zA-Z0-9_-]*/ contained containedin=shComment

highlight default link shdocTag  Keyword
highlight default link shdocArg  Identifier
highlight default link shdocFlag Function
