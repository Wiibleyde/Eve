# Spec: Streamer (Twitch notifications)

## Goal

Notify a channel (with optional role ping) when a tracked Twitch streamer goes live; keep the notification message updated (title/game/viewers) and mark it ended when the stream stops.

## Commands

`/streamer` (guild only, requires Manage Channels):

- `add` — options: `streamer`* (Twitch login), `channel`* (notification channel), `role` (optional ping role).
- `remove` — option: `streamer`* (Twitch login).
- `list` — *(new, TS lacked it)* show tracked streamers for this guild.

On `add`: resolve login → Twitch user ID via Helix; unknown login → ephemeral error. Store the **user ID** (login changes break tracking — TS learned this late, hence its `twitch_id` migration).

## Data model (ent)

`stream`:

- `id`, `guild_id`, `channel_id`, `role_id` (nullable), `twitch_user_id`, `twitch_login` (display/cache), `message_id` (nullable — current live notification), `created_at`.
- Unique on (`guild_id`, `twitch_user_id`).

## Twitch client

`internal/twitch/` package, plain Helix REST:

- App access token via client-credentials, cached until expiry, auto-refresh.
- `GET /helix/users?login=` (resolve), `GET /helix/streams?user_id=&user_id=...` (batched, up to 100 IDs/request).

## Poller

Goroutine ticker, interval **60s** (TS polled every **13s** — needless API pressure; live-notification latency of ≤1min is fine). Skips when maintenance enabled or no rows.

Each tick, diff current live set vs previous in-memory state:

- **Started**: build embed (title, game, viewer count, thumbnail w/ cache-busting query param, link `twitch.tv/<login>`), send to configured channel with role mention (AllowedMentions restricted to that role), store `message_id`.
- **Updated** (title/game changed): edit the message. Skip viewer-count-only edits more often than every 5 min (edit rate discipline).
- **Ended**: edit embed to «Stream terminé» (offline image, duration), clear `message_id`.

On startup, reconcile: fetch live states once, treat already-live streams as "started" only if no `message_id` stored; otherwise resume updating the existing message.

## Env

```
TWITCH_CLIENT_ID=...
TWITCH_CLIENT_SECRET=...
```

Missing → feature disabled (commands not registered, poller not started), warning logged.

## Improvements vs TS

- 60s poll instead of 13s cron.
- Startup reconciliation instead of blind re-notify.
- `list` subcommand.
- AllowedMentions restricted.

## Acceptance

- Add → goes live → notif + ping; title change → edit; stream end → «terminé»; bot restart mid-stream → no duplicate notif.
