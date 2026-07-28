# Spec: Talk

## Goal

Make the bot say something — in the current channel, or as a DM to a user.

## Command

`/talk` — options:

- `message`* (string) — text to send.
- `mp` (user, optional) — if set, send as DM to that user instead of the channel.

**Restriction:** owner-only (`BOT_OWNER_ID`), unlike TS which let anyone puppet the bot — that's an abuse vector (impersonation, spam). Non-owner → ephemeral error.

## Flow

- Channel mode: send `message` as a plain bot message in the interaction channel; ephemeral confirmation «Message envoyé.».
- DM mode: DM the target user; if the DM thread bridge ([mp-threads.md](mp-threads.md)) is implemented, the sent DM is mirrored into the user's MP thread. DM failure (closed DMs) → ephemeral error.
- Sanitize: block `@everyone`/`@here` by sending with disgo `AllowedMentions` = users only.

## Improvements vs TS

- Owner-gated.
- AllowedMentions hardening.

## Acceptance

- Channel + DM paths work; mass mentions never ping.
