# tk

A terminal-first task manager that stores tasks as markdown files. Works alongside Obsidian and git.

Each task is a `.md` file with YAML frontmatter. No database, no sync — just files.

## Install

Requires Go and [fzf](https://github.com/junegunn/fzf).

```
git clone https://github.com/nickhudkins/tk
cd tk
make install
```

Create a config at `~/.config/tk/config.yaml`:

```yaml
root: ~/path/to/your/tasks
```

## Quick start

```
tk add "fix auth redirect loop"       # capture to inbox
tk add "review PR" -d "PR #342"       # with description
tk                                     # dashboard
```

## Status flow

Tasks move through a GTD-inspired pipeline:

```
inbox → todo → next → now → done
```

- **inbox** — raw captures, unprocessed
- **todo** — backlog
- **next** — this week
- **now** — today
- **done** — finished

`tk promote <id>` advances one step. `tk done <id>` jumps straight to done. `tk archive <id>` shelves from any status.

## Commands

| Command | Alias | What it does |
|---------|-------|--------------|
| `tk` | | Dashboard — focus + now + next |
| `tk add "title"` | | New task (inbox) |
| `tk list` | `ls` | List active tasks |
| `tk pick` | `i` | Interactive fzf picker (loops) |
| `tk show <id>` | | View task details |
| `tk edit <id>` | `e` | Open in `$EDITOR` |
| `tk done <id>` | `d` | Mark done |
| `tk promote <id>` | `adv` | Advance status |
| `tk priority <id> <p>` | `p` | Set p0/p1/p2 |
| `tk tag <id> <tag>` | `t` | Add tag |
| `tk archive <id>` | `ar` | Archive |
| `tk delete <id>` | `rm` | Delete permanently |
| `tk today` | `td` | Show today's tasks |
| `tk plan` | | Multi-select tasks → now |
| `tk focus` | | Print focus items |
| `tk next` | `n` | Next action per task |
| `tk search` | `s` | Search / fzf filter |
| `tk review` | | Batch-archive stale tasks |
| `tk export` | | Markdown overview to stdout |

### Flags

- `--json` — structured output (works on list, today, search)
- `--inbox` — inbox only
- `--all` — include done/archived
- `--stale` — stale tasks only
- `--status <s>` — filter by status

## Interactive picker (`tk pick`)

Stays open after each action. Process multiple tasks without re-running.

| Key | Action |
|-----|--------|
| Enter | Edit in `$EDITOR` |
| Tab | Multi-select |
| Ctrl-P | Advance status |
| Ctrl-D | Done |
| Ctrl-A | Archive |
| Ctrl-X | Delete |
| Ctrl-R | Set priority |
| ESC | Exit |

## Task format

```markdown
---
id: 42
title: Fix sidebar toggle bug
status: todo
priority: p0
tags: [code, chromium]
created: 2026-02-23T10:30:00-08:00
updated: 2026-02-23T10:30:00-08:00
---

Notes, links, whatever.

- [ ] First sub-task
- [x] Completed sub-task
```

Sub-tasks use standard markdown checkboxes. `tk next` shows the first unchecked one.

## Typical workflow

```
tk add "thing"          # capture anytime
tk list --inbox         # process inbox
tk promote 42           # inbox → todo

tk plan                 # weekly: pick tasks for today
tk                      # daily: see dashboard
tk pick                 # work through tasks interactively

tk review               # clean up stale stuff
```
