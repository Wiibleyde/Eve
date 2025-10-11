# Eve – Copilot instructions

## Quick start

- Runtime is TypeScript/ESM on Node 22+ with Bun tooling; install with `bun install`, run locally via `bun index.ts`, build distributables with `npm run build` (Bun bundler + `postbuild` copies Prisma assets).
- Prisma targets MySQL/MariaDB; keep `DATABASE_URL` valid and run `npx prisma migrate dev` (or `npm run generate`) after changing `prisma/schema.prisma`.
- Lint with `npm run lint`; format with `npm run format`. Docker images build via multi-stage `Dockerfile` and start through `docker-entrypoint.sh` (runs `prisma migrate deploy` then `node index.js`).

## Core architecture

- `index.ts` wires everything: loads env (`utils/core/config` throws when required vars missing), connects Prisma (`utils/core/database`), inits AI (`utils/intelligence`), loads Discord events, starts the REST API on port 3000, and listens for shutdown signals.
- Discord client lives in `bot/bot.ts`; events registered at startup through `bot/events/event.ts` using an array of handlers in `bot/events/handlers/*`.
- REST endpoints sit under `api/handlers/*`; they lean on shared state from event handlers (e.g. `currentMusic` exposed via `/api/v1/music`).
- Background automation runs from cron jobs in `cron/*.ts`; they are started inside `clientReadyEvent` once the bot logs in.

## Discord surface area

- Slash commands live in `bot/commands/handlers/*` and must be exported via `commands` in `bot/commands/command.ts`; deploys happen automatically in `clientReadyEvent` through `deployCommands`.
- Buttons (`bot/buttons`), modals, and context menus mirror that structure: register handlers in the respective index file so `interactionCreate` can locate them.
- Embed styling is centralized in `bot/utils/embeds.ts`; prefer those helpers (`basicEmbedGenerator`, `successEmbedGenerator`, etc.) over ad‑hoc `EmbedBuilder`s for consistent branding.
- Maintenance mode is toggled by the `/maintenance` command and enforced across handlers via `utils/core/maintenance`; check it before doing heavy work if you add new listeners.

## Data layer & persistence

- Reuse the shared `prisma` client from `utils/core/database`—never instantiate Prisma directly in features.
- Key models: `GlobalUserData` (per-user stats + birthdays), `Config` (guild-scoped key/value settings), `Quote`, `QuizQuestions`, `Stream`, and `LsmsDutyManager`; relationships rely on cascading deletes, so ensure foreign keys exist when inserting records.
- When adding schema fields, update related migrations and refresh generated types with `npx prisma generate`; remember `GuildData.CalendarMessageData` references `BotMessageData` for persistent Discord messages.

## Integrations & services

- AI replies use Google GenAI (`utils/intelligence`); the handler keeps per-channel chat sessions and enforces short 1024-char responses—make sure new features respect `isAiActive`.
- Twitch monitoring lives in `utils/stream/twitch.ts`; the cron job polls every 13s and delegates Discord messaging to `bot/utils/stream.ts`. Update caches via the exported helpers instead of poking at internals.
- MP relay threads are managed in `utils/mpManager.ts`; DMs go through stored thread IDs under the `MP_CHANNEL` guild channel.
- `messageCreate` also houses pattern-based jokes (`utils/messageManager`) and AI mentions—preserve the order of guards when extending it.

## Conventions & utilities

- Project uses module path aliases `@bot/*`, `@utils/*`, `@cron/*` from `tsconfig.json` (bundler resolution). Prefer these imports for cross-package references.
- Strings aimed at users are predominantly French; match tone and language when adding responses.
- Logging goes through the global `logger` (`utils/core/logger`), which mirrors output to Discord via webhook; avoid `console.log` and favor `logger.debug/info/warn/error`.
- Graceful shutdown relies on `endingScripts` (runs LSMS summaries) and `stopBot`; ensure long-running tasks clean up on SIGINT/SIGTERM.

## Adding new behavior

- For new cron tasks, create a module in `cron/` and start it inside `clientReadyEvent`; remember to guard with `isMaintenanceMode` where appropriate.
- Expose new REST endpoints by extending `api/server.ts`; follow the established `/api/v1/...` prefix and keep handlers side-effect free.
- When manipulating streams or persistent Discord messages, update `BotMessageData` / `Stream` rows so reboots via `handleInitStreams` can reconcile state.
- If you introduce additional assets (images, fonts), place them under `assets/` and ensure the build step copies them (see `postbuild`).
