# Spec: Motus (word game)

## Goal

Wordle/Motus-style game in a channel: guess a French word in 6 tries, letters marked found/misplaced/absent.

Depends on: [00-interactions-router.md](00-interactions-router.md).

## Command

`/motus` — starts a game. Public message with game board embed + button «Essayer» (`motus:try`). Ephemeral confirmation to launcher.

## Word source

TS fetched a random French word from an external API (`trouve-mot.fr`). Go: same approach acceptable, but **bundle a fallback word list** (`assets/motus/words.txt`, common French words 6–9 letters) used when the API fails. Normalize: uppercase, strip diacritics.

## Game rules

- Word length: whatever the source returns (typically 6–9). First letter revealed on the initial board (classic Motus).
- 6 attempts max, any user in the channel can try (collaborative, like TS).
- Attempt validation: correct length, letters only, game still running. Invalid → ephemeral error, attempt **not** consumed.
- Letter scoring (two-pass, standard):
  1. Exact position matches → FOUND (red square in Motus convention).
  2. Remaining letters present elsewhere (respecting letter counts) → MISPLACED (yellow circle); else NOT_FOUND.
- Win: all letters FOUND. Lose: 6 failed attempts. On end: reveal word, disable button.

## Board rendering

Embed with one line per attempt using emoji squares/circles (🟥 found / 🟡 misplaced / 🟦 absent) + letter row in code block. Include attempts count `x/6` and participant attribution (who played which try).

## Interaction flow

1. Button `motus:try` → modal `motus:submit` with one text input (the guess).
2. Modal submit → validate → update game state → edit the public board message → ephemeral feedback (WIN/LOSE/CONTINUE/INVALID).

## State

`active_motus` table (same restart-survival rationale as quiz):

- `id`, `message_id` (unique), `channel_id`, `guild_id`, `word`, `attempts` (JSON array of `{word, user_id}`), `state` (playing/won/lost), `started_by`, `created_at`, `expires_at` (24h).

Lazy expiry on interaction; expired → ephemeral «La partie a expiré», reveal word.

## Improvements vs TS

- Games in DB, not `Map<messageId, MotusGame>` in RAM.
- Fallback word list — external API outage doesn't kill the feature.
- Word never logged at INFO (TS logged the answer: `logger.info('Motus word: ', word)`) — DEBUG only.

## Acceptance

- Full game win + lose paths; restart mid-game keeps board playable; diacritics in guesses handled («épée» matches «EPEE»).
