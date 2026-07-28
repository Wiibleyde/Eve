# Spec: Presence Status Rotation

## Goal

Rotating bot presence (activity text), with seasonal variants and a maintenance override.

## Behavior

Goroutine ticker, interval **30s** (TS used 10s — pointless churn; 30s is plenty).

Priority order each tick:

1. **Maintenance enabled** → status `dnd`, activity `Watching "la maintenance"`. Set once, skip rotation until disabled.
2. **Seasonal period** → rotate through the seasonal list.
3. Otherwise → rotate through the default list.

Rotation = simple incrementing index per active list (persisting index across restarts unnecessary).

## Status lists

Default:
- Watching «Regarde les merveilles de ce monde.»
- Listening «Écoute vos instructions.»
- Watching «Regarde les données de mission.»
- Watching «Regarde les étoiles.»

Halloween (Oct 24 → Nov 7):
- Competing «Participe à la préparation des citrouilles. 🎃»
- Watching «Regarde les fantômes... 👻»
- Listening «Spooky Scary Skeletons»
- Playing «Joue à des bonbons ou un sort ! 🍬»

Christmas (Dec 1 → Dec 25):
- Competing «Participe à l'emballage des cadeaux. 🎁»
- Watching «Regarde les lutins. 🧝»
- Listening «Chante des chants de Noël»
- Playing «Joue avec le Père Noël. 🎅»

Period boundaries computed against current year at check time (works across year rollover).

## Implementation

- `internal/bot/features/presence/scheduler.go` — `StartScheduler(client)` like the birthday scheduler.
- Use disgo `client.SetPresence(ctx, gateway.WithWatchingActivity(...))` etc.
- Only send a presence update when the computed status **differs from the last one sent** (avoid redundant gateway calls — TS re-sent every 10s).

## Improvements vs TS

- 30s tick + dedupe instead of blind 10s re-send.
- Maintenance presence set once, not re-set every tick.

## Acceptance

- Presence rotates; toggling maintenance flips to DND immediately (≤1 tick); seasonal lists activate on date.
