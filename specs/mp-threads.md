# Spec: MP Threads (DM ↔ thread bridge)

## Goal

Every DM a user sends to the bot is mirrored into a dedicated thread in a staff channel on the home guild; staff replies in the thread are relayed back to the user's DMs. Lets the owner support users through the bot.

## Trigger / flow

Needs `DirectMessages` intent + `MessageContent`.

### Inbound (user → bot DM)
1. `MessageCreate` where `GuildID == nil`, author not bot.
2. Look up thread for the user; if none, create one:
   - Post «Messages privés avec <@user>» in `MP_CHANNEL`, start a public thread on it named after the user, persist mapping.
3. Repost the DM content into the thread: author header + content + attachments (re-upload by URL) + stickers (as links; sticker re-send is unreliable cross-guild — TS filtered to guild-available stickers, keep that best-effort).

### Outbound (staff → thread)
1. `MessageCreate` in a channel that is a registered MP thread, on the home guild, author not the bot.
2. Relay content + attachments to the user's DM.
3. DM failure (closed DMs) → ⚠️ reply in thread.

## Data model (ent)

`mp_thread`: `id`, `user_id` (snowflake, unique), `thread_id` (snowflake, unique), `created_at`.

In-memory `map[threadID]userID` loaded at startup for the hot path (TS did the same — fine), DB is source of truth.

## Config / env

```
EVE_HOME_GUILD=<guild snowflake>   # guild where MP threads live
MP_CHANNEL=<channel snowflake>     # parent channel for threads
```

Both unset → feature off.

## Edge cases

- Thread archived → unarchive before posting.
- Thread deleted manually → drop the row, recreate on next DM.
- Loop safety: never relay messages authored by the bot itself.

## Improvements vs TS

- Thread auto-recreation on deletion (TS errored).
- Unarchive handling.

## Acceptance

- DM in → thread appears with content; staff message in thread → user gets DM; attachments both ways; closed-DM failure surfaced in thread.
