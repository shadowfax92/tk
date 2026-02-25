<div align="center">

# tk

**A terminal-first task manager backed by markdown files.**

*One command. All your tasks, from inbox to done.*

</div>

Tasks as plain `.md` files with YAML frontmatter — readable in the CLI, Obsidian, and any text editor. No database, no sync, just files you can git track.

- **Quick capture** — `tk add "thing"` drops it in inbox, `--now` or `--next` to skip ahead
- **GTD status flow** — `inbox → todo → next → now → done`
- **Due dates** — `tk add "ship it" --due 5` surfaces tasks automatically when deadlines approach
- **Goals & focus** — track deadlines with progress bars, keep top-of-mind items visible
- **Interactive picker** — fzf-powered loop to process tasks without re-running commands
- **Dashboard** — bare `tk` shows focus, goals, due soon, and today's tasks
- **Agent-friendly** — `--json` output on every list command

---

## Install

Requires Go 1.21+ and [fzf](https://github.com/junegunn/fzf).

```sh
git clone https://github.com/nickhudkins/tk
cd tk
make install
```

Config at `~/.config/tk/config.yaml`:

```yaml
root: ~/path/to/your/tasks
focus_items: 3
due_soon_days: 3
```

## Quick Start

```sh
tk add "fix auth redirect loop"
tk add "ship v2" --due 14 --tags api
tk
```

## Status Flow

```
inbox  →  todo  →  next  →  now  →  done
                                      ↕
                                  archived
```

`tk promote <id>` advances one step. `tk done <id>` jumps to done. `tk archive <id>` shelves from any status.

## Commands

### Tasks

```sh
tk add "title"                  # capture to inbox
tk add "title" --due 5          # due in 5 days
tk add "title" --due 2026-03-15 # due on specific date
tk add "title" --now --tags cli # add to now with tags
tk show 42                      # view task details
tk edit 42                      # open in $EDITOR
tk copy 42                      # copy file path to clipboard
tk delete 42                    # delete permanently
```

### Status & Priority

```sh
tk promote 42                   # advance one step (p)
tk demote 42                    # go back one step (b)
tk done 42                      # mark done (d)
tk archive 42                   # shelve task
tk p0 42                        # set priority (p0/p1/p2)
```

### Views

```sh
tk list                         # all active tasks (ls)
tk list --due                   # tasks with due dates, nearest first
tk list --status next           # filter by status
tk list --stale                 # stale tasks
tk list --sort priority --desc  # sort options: id|created|updated|title|status|priority
tk today                        # today's tasks (t)
tk next                         # next tasks (n)
tk actions                      # next action per task (v)
tk search "auth"                # search by title/body (s)
tk export                       # markdown overview
```

### Interactive

```sh
tk pick                         # fzf picker with actions (i)
tk plan                         # multi-select → move to now
```

### Organize

```sh
tk focus                        # show focus items (f)
tk focus edit                   # edit .focus.md
tk goals                        # show goals (g)
tk goals edit                   # edit .goals.yaml
tk tags                         # list tags with counts
tk tags add 42 code             # tag a task
tk review                       # clean up stale tasks
```

### Flags

| Flag | Description |
|------|-------------|
| `--json` | Structured output |
| `--due` | Filter to tasks with due dates |
| `--inbox` | Inbox only |
| `--all` | Include done/archived |
| `--stale` | Stale tasks only |
| `--status <s>` | Filter by status |
| `--sort <field>` | Sort by id, created, updated, title, status, priority |
| `--desc` | Reverse sort |
| `--show-updated` | Show relative update age |

## Dashboard

Bare `tk` shows your daily view:

```
Focus
  - Ship small, ship often

Goals
  Ship v2 API          12/20 endpoints   ████████░░░░  18 days to go
  Launch blog                                           3 days to go

Due Soon
  1. #42 [todo] Fix auth bug p0 #api     due tomorrow

Now (Mon Feb 24)
  1. #51 Review open PR
  2. #60 Deploy staging

Inbox: 3  Todo: 12  Next: 4  Now: 2  Done: 28
```

## Interactive Picker

`tk pick` opens an fzf picker that **loops** — process your whole list without re-running commands.

| Key | Action |
|-----|--------|
| `Enter` | Edit in $EDITOR |
| `Tab` | Multi-select |
| `Ctrl-P` | Advance status |
| `Ctrl-B` | Demote status |
| `Ctrl-D` | Mark done |
| `Ctrl-O` | Archive |
| `Ctrl-X` | Delete |
| `Ctrl-R` | Set priority |
| `Ctrl-T` | Add tag |
| `Ctrl-F` | Cycle status filter |
| `Ctrl-G` | Cycle tag filter |
| `Esc` | Exit |

## Task Format

Each task lives at `<root>/NNN.md`:

```markdown
---
id: 42
title: Fix sidebar toggle bug
status: todo
priority: p0
tags: [code, chromium]
due: 2026-03-15
created: 2026-02-23T10:30:00-08:00
updated: 2026-02-23T10:30:00-08:00
---

Notes, links, whatever you want here.

- [ ] Write the migration script
- [x] Test with staging data
- [ ] Deploy to production
```

## Goals

Track deadlines and measurable progress in `.goals.yaml`:

```yaml
- title: Ship v2 API
  deadline: 2026-03-15
  metric: endpoints
  current: 12
  target: 20
- title: Launch blog
  deadline: 2026-03-01
```

Edit with `tk goals edit`. Shows on dashboard with progress bars and countdown.

## Config

`~/.config/tk/config.yaml`:

| Key | Default | Description |
|-----|---------|-------------|
| `root` | — | Task storage directory |
| `editor` | `nvim` | Editor for `tk edit` |
| `focus_items` | `3` | Number of focus items shown |
| `due_soon_days` | `3` | Days before due date to surface on dashboard |
| `stale_warn_days` | `28` | Days before stale warning |
| `stale_critical_days` | `56` | Days before critical stale warning |
| `demo` | `false` | Hide dashboard content |

---

> Personal tool built for my own workflow. Feel free to fork and adapt.
