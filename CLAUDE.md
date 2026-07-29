# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code style

**No comments.** Code must be readable on its own — express intent through names, small functions, and explicit types instead of prose. Do not write doc comments on exported symbols, inline comments, trailing comments, or package docs.

The only exceptions are Go directives, which are load-bearing and must be kept:

```go
//go:build ...
//go:generate ...
```

Generated code under `internal/database/ent/` is exempt — ent writes its own comments and the directory is never edited by hand.

When a piece of logic feels like it needs a comment to be understood, rename it or extract it until it doesn't.

## Commands

```bash
go run .                                                    # run the bot
go build ./...                                              # build all packages
go generate ./internal/database/...                         # regenerate ent ORM code after schema changes
go run ./internal/database/migrate_gen.go <name>            # generate versioned migration SQL for schema changes
```

Env vars (`.env` at root, loaded by `godotenv`):

```
DISCORD_TOKEN=...                # required
DATABASE_URL=postgres://...      # required
BOT_OWNER_ID=...                 # owner-only commands (/maintenance, /debug)
API_PORT=3000                    # optional, defaults to 3000
EVE_HOME_GUILD=...               # mpthreads: guild hosting the DM bridge
MP_CHANNEL=...                   # mpthreads: parent channel for DM threads
TWITCH_CLIENT_ID=...             # streamer feature; absent = command hidden
TWITCH_CLIENT_SECRET=...
BLAGUE_API_TOKEN=...             # blague feature; absent = command hidden
MOTUS_API_URL=...                # optional, overrides the word API
MOTUS_WORDS_FILE=...             # optional, overrides the bundled word list
```

Start the database:
```bash
docker compose up -d
```

## Architecture

### Request flow

Discord interaction → `bot.Client` event listener → `router.Router` dispatches by command name → feature handler → ent DB query / REST response.

The router (`internal/bot/router/router.go`) dispatches every interaction type: `OnCommand`, `OnButton`, `OnSelectMenu`, `OnModal`, `OnUserContextMenu`, `OnMessageContextMenu`. It must be wired with `r.Attach(client)` before `client.OpenGateway`.

Components and modals use one custom ID scheme, built with `router.BuildCustomID`:

```
<feature>:<action>[:<data>...]
```

The router matches on the `<feature>:<action>` prefix and passes the remaining segments to the handler as `args []string`. Data segments must not contain `:`; Discord caps custom IDs at 100 characters. Every dispatch passes through a maintenance gate (non-owners get an ephemeral warning while maintenance mode is on) and a panic recovery layer, so one broken handler cannot take the bot down.

### Adding a feature

1. Create `internal/bot/features/<name>/command.go` — define `Commands []discord.ApplicationCommandCreate` and `HandleCommand`.
2. Create `internal/bot/features/<name>/register.go` — `Register(r *router.Router)` calls `r.OnCommand(...)`.
3. Wire in `internal/bot/bot.go`: call `feature.Register(r)` and append `feature.Commands...` to `allCommands`. Commands are registered globally on Discord via `SetGlobalCommands` in the `Ready` handler.

Features gated behind an optional env var (blague, streamer, maintenance, talk) expose `Commands()` as a function instead of a variable, returning an empty slice when the credential is missing so the command never shows up.

### Conventions

- User-facing strings are **French**.
- No hardcoded guild, channel, or role IDs in code — they live in the `guild_configs` table or in env vars.
- No hardcoded Discord CDN attachment URLs (they expire) — bundle images under `assets/` or use a stable URL.
- Ephemeral replies go through `helpers.RespondEphemeral*`, message bodies through `internal/bot/ui`.
- Config keys are declared once in `internal/bot/features/config/keys.go`; `/config list`, `get`, and `reset` are generated from that registry.

### Database / ORM

Uses **ent** (entgo.io) with versioned Atlas migrations.

- Schemas defined in `internal/database/tables/` as ent schema structs.
- Generated code lives in `internal/database/ent/` — **do not edit manually**.
- After editing a schema: run `go generate ./internal/database/...` to regenerate, then `go run ./internal/database/migrate_gen.go <migration_name>` to produce the SQL diff.
- `migrate_gen.go` uses `ModeInspect` against `DATABASE_URL` (live DB) — no separate dev DB needed.
- `database.Default.Migrate(ctx)` (called on startup) applies pending migrations via ent's auto-migrate.
- SQL queries log at DEBUG level automatically.

### UI (Components V2)

Every message the bot sends uses Discord **Components V2** — no `discord.Embed` anywhere. `internal/bot/ui/` builds them:

- `ui.New()`, `ui.Success(msg)`, `ui.Error(msg)`, `ui.Warning(title, msg)` return a `*ui.Card`.
- A `Card` renders as one `ContainerComponent` (accent-coloured left border) holding a `## ` title, text blocks, separators, media galleries, action rows, and a `-# ` footer with a timestamp.
- Chainable: `Title`, `Text`/`Textf`, `Heading`, `Subtext`/`Subtextf`, `Fields(...ui.Field)`, `Divider`, `Space`, `Thumbnail`, `Image`/`Images`, `Row(buttons...)`, `Accent`, `Footer`, `NoFooter`.
- Emit with `card.MessageCreate()`, `card.EphemeralCreate()`, `card.MessageUpdate()`, or `helpers.RespondEphemeralCard` / `RespondCard` / `FollowupEphemeralCard` / `EditResponseCard`.

**Mentions ping for real.** Container text is regular message content, unlike embed text, so `<@id>` and `<@&id>` inside a card notify. Every constructor sets `AllowedMentions: ui.NoMentions()`; only override it deliberately (the Twitch poller does, to ping the configured role once on go-live).

Budget: 4000 characters of text and 40 components per message — truncate long lists (`ui.MaxText`).

Reading text back out of a V2 message (e.g. to republish it) goes through `ui.Texts(msg.Components)`.

### Scheduler

`birthday.StartScheduler(client)` runs a goroutine that wakes at midnight, queries today's birthdays, and calls `sendBirthdayMessage` for each. Guild birthday channels are stored in `guild_configs` table with key `"birthday.channel"` (value = channel snowflake ID).
