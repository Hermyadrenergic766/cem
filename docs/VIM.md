# cem in Vim / Neovim

No plugin needed — cem is just a shell command.

---

## Vim

Add to `.vimrc`:

```vim
" Send visual selection to cem (thinker), show output in a buffer
function! CemThink() range
  let l:lines = getline(a:firstline, a:lastline)
  let l:prompt = join(l:lines, "\n")
  silent execute "vsplit | enew | setlocal buftype=nofile"
  silent execute "0read !cem " . shellescape(l:prompt)
endfunction

" Same, but writer mode
function! CemWrite() range
  let l:lines = getline(a:firstline, a:lastline)
  let l:prompt = join(l:lines, "\n")
  silent execute "vsplit | enew | setlocal buftype=nofile"
  silent execute "0read !cem -w " . shellescape(l:prompt)
endfunction

" Same, but pair mode
function! CemPair() range
  let l:lines = getline(a:firstline, a:lastline)
  let l:prompt = join(l:lines, "\n")
  silent execute "vsplit | enew | setlocal buftype=nofile"
  silent execute "0read !cem -p " . shellescape(l:prompt)
endfunction

" Bindings: <Leader>ct / cw / cp
vnoremap <silent> <Leader>ct :call CemThink()<CR>
vnoremap <silent> <Leader>cw :call CemWrite()<CR>
vnoremap <silent> <Leader>cp :call CemPair()<CR>
```

Select lines, press `\ct` / `\cw` / `\cp` (assuming `<Leader>` is backslash). Result appears in a new vertical split.

---

## Neovim with floating window

```lua
local function cem_run(mode)
  local lines = vim.api.nvim_buf_get_lines(0,
    vim.fn.line("'<") - 1, vim.fn.line("'>"), false)
  local prompt = table.concat(lines, "\n")
  local flag = ({ think = "", write = "-w", pair = "-p" })[mode]
  local args = vim.tbl_filter(function(a) return a ~= "" end, { flag, prompt })

  -- Floating window
  local buf = vim.api.nvim_create_buf(false, true)
  local w = math.floor(vim.o.columns * 0.7)
  local h = math.floor(vim.o.lines * 0.7)
  vim.api.nvim_open_win(buf, true, {
    relative = "editor", row = math.floor((vim.o.lines - h) / 2),
    col = math.floor((vim.o.columns - w) / 2), width = w, height = h,
    border = "rounded", title = "cem " .. mode,
  })

  vim.fn.jobstart({ "cem", unpack(args) }, {
    stdout_buffered = true,
    on_stdout = function(_, out) vim.api.nvim_buf_set_lines(buf, -1, -1, false, out) end,
    on_stderr = function(_, err) vim.api.nvim_buf_set_lines(buf, -1, -1, false, err) end,
  })
end

vim.keymap.set("v", "<leader>ct", function() cem_run("think") end)
vim.keymap.set("v", "<leader>cw", function() cem_run("write") end)
vim.keymap.set("v", "<leader>cp", function() cem_run("pair") end)
```

---

## Troubleshooting

### `:!cem` works but mapping doesn't

`<Leader>` may not be what you expect. Run `:echo mapleader` (Vim) or check `vim.g.mapleader` (Neovim).

### Output truncated

`!cem` runs synchronously and blocks Vim. For long responses, switch to async (jobstart in Neovim, `+job_start` in Vim 8+).
