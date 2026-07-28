# Eve Go v2 — Feature Specs

Specs for porting missing features from the TypeScript version (`origin/main`) to the Go rewrite.
These are **redesigns**, not 1:1 ports — known TS bad practices are fixed (noted per spec).

LSMS and Labo features are intentionally dropped.

## Build order

Infrastructure first — most features depend on it:

1. [00-interactions-router.md](00-interactions-router.md) — buttons, modals, select menus, context menus, custom ID scheme
2. [01-maintenance.md](01-maintenance.md) — global gate used by crons/commands
3. [02-presence-status.md](02-presence-status.md) — presence rotation scheduler

Then features (independent of each other):

- [birthday-config-gaps.md](birthday-config-gaps.md) — missing subcommands in existing features
- [quiz.md](quiz.md)
- [motus.md](motus.md)
- [blague.md](blague.md)
- [coinflip.md](coinflip.md)
- [talk.md](talk.md)
- [streamer.md](streamer.md) — Twitch notifications
- [calendar.md](calendar.md) — ICS calendar + Discord scheduled events
- [loto.md](loto.md)
- [message-reactions.md](message-reactions.md) — pattern responses ("quoi" → "Feur.")
- [mp-threads.md](mp-threads.md) — DM ↔ thread bridge
- [context-menus.md](context-menus.md) — quote from message, avatar, banner
- [debug.md](debug.md)
- [api.md](api.md) — missing REST endpoints

## Conventions (all specs)

- User-facing strings: **French** (same as TS version).
- All ephemeral replies via `helpers.RespondEphemeral*`, embeds via `internal/bot/embeds`.
- New feature = `internal/bot/features/<name>/` with `command.go` + `register.go`, wired in `bot.go` (see CLAUDE.md).
- New tables = ent schema in `internal/database/tables/`, then `go generate` + `migrate_gen.go`.
- No hardcoded guild IDs / channel IDs / role IDs in code — everything in `guild_configs` or env.
- No hardcoded Discord CDN attachment URLs (they expire) — bundle images in `assets/` or use stable URLs.
