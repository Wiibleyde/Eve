# Spec: Message Reactions (pattern replies)

## Goal

Joke auto-replies to message patterns — the classic «quoi ?» → «Feur.» plus siblings.

## Trigger

`MessageCreate` (guild messages, not bots, not when the bot is mentioned).

Per-guild opt-out flag `jokes.disabled` in `guild_configs` (**not** the TS hardcoded `jokeIgnoredServers` list). Maintenance → silent skip.

## Detectors

Table-driven, each detector = regex list + weighted responses:

| Detector | Patterns (end-of-message, case-insensitive, repeated letters tolerated) | Responses (weight) |
|---|---|---|
| quoi | `quoi`, `koa`, `qoa`, `koi`, `kwa`, `kewa` (+ trailing spaces/`?`) | Feur. (70), coubeh. (10), la 🐨 (10), drilatère. (10) |
| comment | `comment`, `komen` | dant. (default) |

(Port the exact TS regex lists; they tolerate letter repetition like `quoiiii`.)

Message normalized (lowercase, trim) before matching. First matching detector wins. Weighted pick sums to 100; fall back to detector default.

## Rate limiting

New (TS had none — spam loop risk): per-channel cooldown **30s** between joke replies. In-memory map, no persistence needed.

## Implementation

`internal/bot/features/reactions/` — pure function `Detect(content string) (reply string, ok bool)` + thin event glue. Detectors in one table so adding a new one is a data change.

## Improvements vs TS

- Guild opt-out in DB.
- Channel cooldown.
- Detector logic pure + unit-testable.

## Acceptance

- «quoiiii ??» triggers weighted reply; cooldown suppresses spam; disabled guild silent; unit tests cover regex set.
