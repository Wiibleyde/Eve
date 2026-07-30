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

Start the database (dev — bot itself runs on the host with `go run .`):
```bash
docker compose up -d
```

### Production

`Dockerfile` (multi-stage, `CGO_ENABLED=0`, distroless-ish alpine runtime, non-root uid 10001) + `docker-compose.prod.yml` (bot + postgres, restart policy, healthchecks, log rotation). Dev stays host-native: `docker-compose.yml` is untouched and still only starts postgres.

```bash
cp .env.prod.example .env.prod                                          # then fill in secrets
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f eve
docker compose -f docker-compose.prod.yml --env-file .env.prod down
```

`--env-file` drives compose interpolation, `env_file:` injects the same file into the container — both read `.env.prod`, so the flag is required.

`DATABASE_URL` is derived from `POSTGRES_*` and points at the `postgres` service; set it explicitly in `.env.prod` to use an external database. `TZ` (default `Europe/Paris`) matters — the birthday scheduler wakes at local midnight. Assets are copied to `/app/assets` and the workdir is `/app`, so the relative asset paths in `quote` and `motus` resolve. Version is stamped via `-ldflags -X Eve/internal/version.Version=${VERSION}`.

`IMAGE`/`VERSION` in `.env.prod` select the image; the default `ghcr.io/wiibleyde/eve:latest` is what CI publishes. Drop `--build` and run `pull` + `up -d` to deploy a published image instead of building on the host.

### CI/CD

`.github/workflows/ci.yml` (push to main, PRs, manual) runs five parallel jobs: **style** (gofmt, `go vet`, `.github/scripts/check-no-comments.sh`), **lint** (golangci-lint v2, config in `.golangci.yml`), **test** (`go build ./...`, `go test -race -shuffle=on ./...`), **ent-codegen** (regenerates and fails on drift), **docker** (multi-arch build, no push).

`.github/workflows/build-docker.yml` (push to main only) builds the image and pushes it to `ghcr.io/wiibleyde/eve` as `latest` and `<unix-timestamp>`. No git tags or releases are involved — a commit on main is the only thing that publishes an image. **It never deploys** — pull and restart on the host yourself. It passes no `VERSION` build-arg, so published images report the Dockerfile default (`dev`).

The no-comments rule is enforced in CI by a grep over `*.go` that exempts `//go:build|generate|embed` and `internal/database/ent/`.

## Architecture

### Request flow

Discord interaction → `bot.Client` event listener → `router.Router` dispatches by command name → feature handler → ent DB query / REST response.

The router (`internal/bot/router/router.go`) dispatches every interaction type: `OnCommand`, `OnButton`, `OnSelectMenu`, `OnModal`, `OnUserContextMenu`, `OnMessageContextMenu`. It must be wired with `r.Attach(client)` before `client.OpenGateway`.

Components and modals use one custom ID scheme, built with `router.BuildCustomID`:

```
<feature>:<action>[:<data>...]
```

The router matches on the `<feature>:<action>` prefix and passes the remaining segments to the handler as `args []string`. Data segments must not contain `:`; Discord caps custom IDs at 100 characters. Every dispatch passes through a maintenance gate (non-owners get an ephemeral warning while maintenance mode is on) and a panic recovery layer, so one broken handler cannot take the bot down.

### REST API / OpenAPI

`internal/api/server.go` mounts a Fiber v3 app on `API_PORT`. Swagger UI is served at `/docs`, the spec at `/openapi.yaml`. The API is GET-only (CORS enforces it).

The `routes()` table in `server.go` is the single source of truth: it declares path, handler, tag, operation ID, summary, query shape and response codes. `Start` loops over it to register the Fiber routes **and** passes it to `buildSpec`, so a route and its documentation cannot drift apart. Adding an endpoint means adding one row.

The spec is **not** a checked-in file and **not** generated from annotations — swaggo/swag needs comments, which CI rejects. `internal/api/openapi.go` builds it at startup with `swaggest/openapi-go`, reflecting the real structs:

- Schemas come from the response types themselves (`models.*Response`, `controllers.Loto*Response`, `controllers.APIError`).
- Field prose lives in struct tags: `description`, `example`, `format`, `nullable`, `enum`, `minimum`, `maximum`, `default`. `nullable:"false"` on a slice suppresses the null that reflection adds by default.
- Query parameters are the `statsQuery` / `winnersQuery` structs (`query:"..."` tags). Handlers still read `c.Query(...)` directly, so these describe the surface rather than bind it — keep them in sync when a parameter changes.
- Two reflector hooks: `unprefixedDefName` strips the package prefix from schema names (`ModelsHealthResponse` → `HealthResponse`), `requireResponseFields` marks response fields required without making query parameters required.

`buildSpec` returns an error rather than panicking; `Start` logs it and skips mounting the UI, so a broken spec can never stop the API or the bot. `openapi_test.go` asserts every declared route reaches the spec.

Swagger UI's JS/CSS load from the unpkg CDN, so the `/docs` page needs internet access in the browser. The spec itself is served from memory.

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
