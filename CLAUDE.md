# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go run .                                                    # run the bot
go build ./...                                              # build all packages
go generate ./internal/database/...                         # regenerate ent ORM code after schema changes
go run ./internal/database/migrate_gen.go <name>            # generate versioned migration SQL for schema changes
```

Required env vars (`.env` at root, loaded by `godotenv`):
```
DISCORD_TOKEN=...
DATABASE_URL=postgres://...
```

Start the database:
```bash
docker compose up -d
```

## Architecture

### Request flow

Discord interaction → `bot.Client` event listener → `router.Router` dispatches by command name → feature handler → ent DB query / REST response.

The router (`internal/bot/router/router.go`) is a thin dispatcher: `OnCommand(name, fn)` maps slash command names to handlers. It must be wired with `r.Attach(client)` before `client.OpenGateway`.

### Adding a feature

1. Create `internal/bot/features/<name>/command.go` — define `Commands []discord.ApplicationCommandCreate` and `HandleCommand`.
2. Create `internal/bot/features/<name>/register.go` — `Register(r *router.Router)` calls `r.OnCommand(...)`.
3. Wire in `internal/bot/bot.go`: call `feature.Register(r)` and append `feature.Commands...` to `allCommands`. Commands are registered globally on Discord via `SetGlobalCommands` in the `Ready` handler.

### Database / ORM

Uses **ent** (entgo.io) with versioned Atlas migrations.

- Schemas defined in `internal/database/tables/` as ent schema structs.
- Generated code lives in `internal/database/ent/` — **do not edit manually**.
- After editing a schema: run `go generate ./internal/database/...` to regenerate, then `go run ./internal/database/migrate_gen.go <migration_name>` to produce the SQL diff.
- `migrate_gen.go` uses `ModeInspect` against `DATABASE_URL` (live DB) — no separate dev DB needed.
- `database.Default.Migrate(ctx)` (called on startup) applies pending migrations via ent's auto-migrate.
- SQL queries log at DEBUG level automatically.

### Embeds

`internal/bot/embeds/` has three builders: `BaseEmbed()`, `SuccessEmbed(msg)`, `ErrorEmbed(msg)`. All responses to interactions should use `helpers.RespondEphemeral*` from `internal/bot/helpers/responses.go`.

### Scheduler

`birthday.StartScheduler(client)` runs a goroutine that wakes at midnight, queries today's birthdays, and calls `sendBirthdayMessage` for each. Guild birthday channels are stored in `guild_configs` table with key `"birthday.channel"` (value = channel snowflake ID).
