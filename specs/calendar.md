# Spec: Calendar (ICS)

## Goal

Per-guild ICS calendar: a persistent auto-refreshing embed listing current/upcoming events, plus automatic creation of Discord Scheduled Events shortly before each event starts.

## Commands

`/calendar` (guild only, requires Manage Channels):

- `set` — option `url`* (ICS URL). Validates by fetching+parsing before saving. Posts the calendar embed in the current channel and stores its message ref.
- `remove` — deletes config + the calendar message (best effort).
- `refresh` — force-refresh the embed now.

TS also had `/calendar-test` (debug dump of parsed events) — fold into `refresh` output or drop; **drop** (owner can read logs).

## Data model (ent)

Extend `guild_configs` (existing key/value) — no new table needed:

- `calendar.url` — ICS URL
- `calendar.channel` — channel ID of the embed message
- `calendar.message` — message ID of the embed message

(TS used a `GuildData` row + a separate `BotMessageData` relation — overkill for one message ref.)

## ICS parsing

Use a Go ICS lib (e.g. `github.com/arran4/golang-ical`) or minimal hand parser for VEVENT `DTSTART`/`DTEND`/`SUMMARY`/`DESCRIPTION`/`LOCATION`/`UID`.

Must handle: timezones (TZID), all-day events. Recurring events (RRULE): out of scope v1 — document limitation.

## Embed

- «📅 Calendrier» title.
- Section «En cours» (events where start ≤ now ≤ end) and «À venir» (next 5, sorted by start), each line: `**summary** — <t:start:R>` + location if present.
- Footer: last refresh timestamp.

## Scheduler

Single goroutine, tick every **5 min** (same as TS), skip in maintenance:

1. For each guild with `calendar.url`:
   - Fetch + parse ICS (per-guild errors logged, don't abort the loop).
   - Edit the stored message with the fresh embed. Message deleted/channel gone → clear the stored refs, log warning.
2. Events starting within **30 min**: create a Discord Scheduled Event (external type, name/description/start/end/location) if one with same name+start doesn't already exist. Dedupe by querying the guild's scheduled events (TS also kept an in-memory set — redundant with the query, drop it).

## Improvements vs TS

- Config in `guild_configs` keys, no bespoke tables.
- Dedupe scheduled events purely by API query — restart-safe.
- Per-guild error isolation in the cron loop.

## Acceptance

- Set URL → embed appears + auto-refreshes; event <30min away → Discord scheduled event created once; remove cleans up.
