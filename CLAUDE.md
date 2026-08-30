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
GOOGLE_API_KEY=...               # AI feature; absent = mention replies disabled
LAVALINK_ADDRESS=host:port       # music feature; absent = commands hidden
LAVALINK_PASSWORD=...
LAVALINK_SECURE=true             # optional, reach the node over https/wss
YTDLP_PATH=/usr/bin/yt-dlp       # optional, defaults to yt-dlp on PATH
YTDLP_JS_RUNTIME=node            # optional, yt-dlp --js-runtimes (default: deno)
```

Start the database and Lavalink (dev — bot itself runs on the host with `go run .`):
```bash
docker compose up -d
```

`docker compose up -d` starts postgres and lavalink (published on `2333`). Set `LAVALINK_ADDRESS=localhost:2333` in `.env` for the host-native bot to reach it. The AI feature needs no local service — it just needs `GOOGLE_API_KEY` in `.env`.

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

### AI (Gemini)

`internal/gemini` is a thin wrapper over the official `google.golang.org/genai` SDK, `internal/bot/features/ai` is the Discord side. `GOOGLE_API_KEY` gates the whole thing — unset means the listener is never attached. This mirrors the old TypeScript bot's `utils/intelligence.ts`, not the CPU-bound Ollama setup it briefly replaced: `gemini-3.7-flash`, one persistent chat session per channel, and the `googleSearch` tool doing grounding provider-side instead of a hand-rolled SearXNG lookup.

`gemini.Client.NewChat` creates a `genai.Chat` with safety settings on `BLOCK_NONE` across all harm categories and `Tools: [{GoogleSearch: {}}]` — the model decides on its own when a question needs a web search, no local heuristic involved. The SDK owns the conversation history inside the `*genai.Chat` value; Eve never sees or stores message content herself.

`ai.Attach` replies when the bot is @-mentioned or when a user replies to one of its messages, never to `@everyone`, bots, webhooks, DMs (that is `mpthreads`), or while maintenance mode is on. Generation runs in a goroutine behind a panic recovery, with a typing indicator refreshed every 8s.

Context is per channel and in memory only (`sessions.go`): one `genai.Chat` per channel, dropped after 30 minutes idle, gone on restart. `sessionStore` only tracks what Eve needs beyond the SDK's own history — the channel's participant set (for the mention allow-list below) and the idle TTL. User turns are sent as `<display name> (<@id>) : <text>` so the model can tell speakers apart and mention them back.

`prompt.go` builds the system instruction once per new chat session, because it embeds runtime IDs: the bot's own ID (so Eve never mentions herself) and `BOT_OWNER_ID` (the creator sentence is dropped when no owner is configured). The persona is the WALL-E scout robot — direct, warm in social settings, French, 1024 characters max — plus the seven `<:eve*:>` emojis declared in `emojis.go`. Gemini is far more capable than the old 1B local model, so `cleanReply`'s 1024-character truncation is the only hard limit left, not the prompt.

**Eve's replies ping for real, and the model decides who.** That is the point (the prompt tells her to mention whoever she answers), so it is fenced in three layers rather than trusted:

- `cleanReply` deletes self-mentions, role mentions (`<@&…>`) and half-written placeholders (`<@ID du compte>`, `@ID du compte`), and defangs `@everyone`/`@here`.
- `AllowedMentions.Parse` is left empty and `Users` is set to an explicit ID list, so anything that survives step 1 still renders as inert text unless it is on that list.
- The list is `pingable(content, allowedTargets(...))`: only IDs already in the conversation — channel participants, the current author, the owner — can be pinged. A user who talks Eve into mentioning a stranger gets text, not a notification.

`RepliedUser` is only true when the reply contains no ping of its own, so the author is never notified twice.

Per-guild opt-out is the `ai.disabled` config key (`/config set ai-disabled`). A database outage makes the bot stay silent rather than answer unfiltered.

Gemini's own rate limiting surfaces as a `*genai.APIError` with `Code` 429 or 503; `gemini.translateError` maps both to the sentinel `gemini.ErrRateLimited`, which `respondError` turns into the "trop de demandes" card instead of a generic failure.

### Music (yt-dlp + Lavalink)

`internal/audio` is the audio backend, `internal/bot/features/music` is the Discord side. `LAVALINK_ADDRESS` gates the whole thing — unset means `music.Commands()` returns nothing and the thirteen slash commands never register.

**Lavalink does not extract, it only plays.** Its `youtube-plugin` broke on this deployment (`Must find sig function from script` on WEB, `This video requires login` on ANDROID_VR), and that class of breakage recurs every time YouTube ships a new player. So extraction is `yt-dlp` (`internal/audio/ytdlp.go`) and Lavalink receives a plain `http` URL. The node keeps doing voice, opus, filters and seeking; it just never talks to YouTube.

The consequence drives the whole design: **googlevideo URLs are IP-bound and expire in a few hours**, so the queue cannot hold resolved tracks. It holds `audio.Media` (title, author, URI, artwork, length) and `startPlayback` resolves a fresh stream URL through `audio.Stream` at the moment a track actually starts. `lavalink.Track` exists only long enough to be handed to `player.Update`.

That also means `TrackStartEvent`/`TrackEndEvent` carry a useless `http` track titled `Unknown title`. Every card reads `guildState.current()` instead, which is why the state owns a `playing *audio.Media`.

`yt-dlp` needs a JavaScript runtime for YouTube now. The image installs `nodejs` and sets `YTDLP_JS_RUNTIME=node`; unset means yt-dlp's default (deno) and a warning if it is absent. The runtime image installs yt-dlp from pip, not apk — Alpine's package was nine months stale, and a stale yt-dlp fails exactly like the plugin it replaced. It needs rebuilding periodically.

`audio.New` spawns `AddNode` in a goroutine because `node.Open` retries forever until it connects. Every handler goes through `ready()`, which fails with a card when no node is connected yet.

Discord never streams audio through the bot process: `bot.Client` only forwards `GuildVoiceStateUpdate` and `VoiceServerUpdate` to disgolink. That needs `gateway.IntentGuildVoiceStates`, and `cache.FlagVoiceStates` is enabled so `/play` can find the caller's voice channel — the only two entries in an otherwise empty cache config.

Queue, repeat mode, history, the now-playing message ID and the lyrics thread live in `state.go`, per guild, **in memory only** — a restart drops them while the node keeps playing, so playback is orphaned until someone runs `/stop`.

The now-playing message is a single card edited in place on every `TrackStartEvent`, carrying `music:back`, `music:skip`, `music:playpause`, `music:loop`. Advance is driven by `TrackEndEvent` and `Reason.MayStartNext()`. A track that fails to resolve is skipped rather than retried, so a dead URL cannot wedge the queue — and `play` never re-enters itself on `RepeatTrack` after a failure.

The bot leaves after 60s idle (`leaveOnEndDelay`) or 60s alone in the channel (`leaveOnEmptyDelay`), both cancellable timers on the guild state.

`/filter` cannot mirror the old discord-player list — Lavalink exposes a fixed set, so `filters.go` declares named presets over it. `/loop` has no autoplay mode: nothing in Lavalink recommends a next track.

`/syncedlyrics` uses `dev.schlaubi.lyrics:lavalink` (lyrics.kt). It is **not** LavaLyrics — different route, different JSON, and **no subscribe or websocket line events**, so the bot drives the timing itself: fetch the timed lines once, then a goroutine ticks every 500ms and posts each line as `player.Position()` passes its `range.start`. It stops on context cancel, on the current track changing, or when the lines run out.

Lyrics are looked up by **video ID**, not by the playing track — the node only sees an anonymous `http` URL, so the player-scoped route would always 404. `audio.VideoID` pulls the ID out of the stored `Media.URI`, and a search through `/v4/lyrics/search` is the fallback when there is no ID or no match.

That fallback is where wrong lyrics come from, so `match.go` guards it. Feeding the raw YouTube title into the search returns garbage: `Clair Obscur: Expedition 33 | Lumière [Official Music Video] Sandfall Interactive` ranks *Une vie à t'aimer* (by a band called Clair Obscur) first. So the query is built from a real artist/title pair — `yt-dlp`'s `track`/`artist` tags when the video carries them, otherwise `splitArtistTitle` splitting on ` - ` / ` | ` after stripping bracketed noise — and every candidate must clear `containment` ≥ 0.75 against the expected title, both in the search result and again in the returned `track.title`. Nothing clears it, no lyrics: a wrong match is worse than none. Matching is accent- and punctuation-insensitive with filler words dropped. `match_test.go` pins the Lumière case.

### Scheduler

`birthday.StartScheduler(client)` runs a goroutine that wakes at midnight, queries today's birthdays, and calls `sendBirthdayMessage` for each. Guild birthday channels are stored in `guild_configs` table with key `"birthday.channel"` (value = channel snowflake ID).
