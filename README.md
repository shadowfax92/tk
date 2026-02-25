<div align="center">

# ✅ tk

**A terminal-first task manager backed by markdown files.**

*One command. All your tasks, from inbox to done.*

</div>

You have tasks scattered across apps, notes, and your head. tk stores each task as a plain `.md` file with YAML frontmatter — readable in the CLI, Obsidian, and any text editor. No database, no sync, just files you can git track.

- 📥 **Quick capture** — `tk add "thing"` drops it in inbox/next/now, zero friction
- 🔄 **GTD status flow** — tasks move through `inbox → todo → next → now → done`
- 🔍 **Interactive picker** — fzf-powered loop to process tasks without re-running commands
- 📋 **Daily dashboard** — bare `tk` shows your focus, today's tasks, and what's next
- 🏷️ **Tags & priorities** — lightweight categorization without imposed project structure
- 🤖 **Agent-friendly** — `--json` output on every list command for programmatic use

---

## Install

Requires Go 1.21+ and [fzf](https://github.com/junegunn/fzf).

```sh
git clone https://github.com/nickhudkins/tk
cd tk
make install    # builds and copies to $GOPATH/bin/
```

Create a config at `~/.config/tk/config.yaml`:

```yaml
root: ~/path/to/your/tasks
focus_items: 3
demo: false
```

That's it. `tk add "first task"` to get started.

## Quick Start

```sh
# 1. Capture tasks
tk add "fix auth redirect loop"
tk add "review PR" -d "needs security review"

# 2. See your dashboard
tk
```

## Status Flow

Tasks move through a GTD-inspired pipeline that naturally narrows your focus:

```
📥 inbox  →  📋 todo  →  🔜 next  →  🔥 now  →  ✅ done
                                                    ↕
                                                📦 archived
```

- **inbox** — raw captures, unprocessed
- **todo** — categorized backlog
- **next** — planned for this week
- **now** — today's focus
- **done** — completed

`tk promote <id>` advances one step. `tk done <id>` jumps straight to done. `tk archive <id>` shelves from any status.

## CLI

```sh
tk                              # dashboard — focus + now + next
tk add "title"                  # capture to inbox
tk add "title" -d "details"     # with description
tk add "title" --next           # add directly to next
tk add "title" --now            # add directly to now
tk list                         # list active tasks
tk list --status done --sort updated --show-updated   # most recent done first, with updated age
tk demo on                      # hide dashboard content during screenshare
tk pick                         # interactive fzf picker (loops)
tk show 42                      # view task details
tk edit 42                      # open in $EDITOR
tk done 42                      # mark done
tk promote 42                   # advance status one step
tk priority 42 p0               # set priority (p0/p1/p2)
tk p0 42                        # quick set priority p0 (also p1/p2)
tk tag 42 code                  # add a tag
tk archive 42                   # shelve task
tk delete 42                    # delete permanently
tk today                        # show 'now' tasks
tk plan                         # fzf multi-select → move to now
tk focus                        # print random focus items
tk next                         # show 'next' tasks
tk actions                      # next action per now task
tk actions --next               # next action per next task
tk search "auth"                # search tasks
tk review                       # batch-archive stale tasks
tk export                       # markdown overview to stdout
```

### Flags

- `--json` — structured output (works on list, today, search)
- `--inbox` — inbox items only
- `--all` — include done/archived
- `--stale` — stale tasks only
- `--status <s>` — filter by status
- `--sort <field>` — sort list by `id|created|updated|title|status|priority` (smart defaults per field)
- `--desc` — reverse sort order for the selected field
- `--show-updated` — append colored relative updated age (for example, `today`, `1 day ago`, `2 weeks ago`)

Config:
- `focus_items` — number of focus items to show (randomly sampled each run for `tk` and `tk focus`)
- `demo` — when true, bare `tk` prints only `tk demo mode`

## Interactive Picker

`tk pick` (alias `tk i`) opens an fzf picker that **loops** — after each action, it reloads and re-enters fzf. Process your whole list without re-running commands.

| Key | Action |
|-----|--------|
| `Enter` | Edit in `$EDITOR` |
| `Tab` | Multi-select |
| `Ctrl-P` | Advance status |
| `Ctrl-D` | Mark done |
| `Ctrl-A` | Archive |
| `Ctrl-X` | Delete |
| `Ctrl-R` | Set priority |
| `Esc` | Exit picker |

## Task Format

Each task lives at `<root>/NNN.md`:

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

Notes, links, whatever you want here.

- [ ] Write the migration script
- [x] Test with staging data
- [ ] Deploy to production
```

Sub-tasks use standard markdown checkboxes. `tk actions` shows the first unchecked one per task.

## Workflow

**📥 Capture** (anytime)
```sh
tk add "fix the auth redirect loop"
tk add "review open PR" -d "needs security review"
```

**📋 Process inbox** (periodic)
```sh
tk list --inbox
tk promote 42       # inbox → todo
tk delete 43        # not useful
```

**📅 Plan the week**
```sh
tk plan             # fzf multi-select from todo/next → now
```

**🔥 Daily execution**
```sh
tk                  # dashboard
tk pick             # work through tasks interactively
```

**🧹 Clean up**
```sh
tk review           # batch-archive stale tasks
```

## Dashboard

Bare `tk` shows your daily view:

```
Focus
  - Be nice

Now (Mon Feb 24)
  1. #42 Fix auth redirect loop
  2. #51 Review open PR

Next
  #53 Write CDP test harness
  #55 Update deployment docs

Inbox: 3  Todo: 12  Next: 4  Now: 2  Done: 28
```

---

> This is a personal tool built for my own workflow. Sharing it in case it's useful — feel free to fork and adapt.
