# Spec: Maintenance Mode

## Goal

Global switch that pauses user-facing bot activity. Used as a gate by crons (stream poller, calendar, status rotation) and interaction handlers.

## Command

`/maintenance` — toggle. **Owner-only** (checked against `BOT_OWNER_ID` env var, not a Discord permission).

- Non-owner → ephemeral error.
- On enable: presence switches to DND + "Regarde la maintenance" (handled by presence scheduler, see [02-presence-status.md](02-presence-status.md)); ephemeral confirmation.
- On disable: normal presence rotation resumes; ephemeral confirmation.

## Behavior while enabled

- All slash/component/modal interactions from non-owner users → ephemeral warning embed: "Le bot est actuellement en mode maintenance. Veuillez réessayer plus tard."
  - Implement as a **check in the router dispatch layer** (single place), not per-handler like TS.
- Owner interactions pass through normally.
- Schedulers (stream, calendar, birthday, status except maintenance presence) skip their tick.
- `messageCreate`-driven features (AI chat, pattern replies) reply with the same warning only when directly addressed (mention); pattern replies just stay silent.

## Implementation

```go
// internal/bot/maintenance/maintenance.go
package maintenance

func Enabled() bool
func Set(v bool)
func Toggle() bool
```

Atomic bool (`sync/atomic`), in-memory only — state resets on restart, which is acceptable (restart = end of maintenance).

## Env

```
BOT_OWNER_ID=<discord user snowflake>   # new, add to config.Config
```

## Improvements vs TS

- TS checked `isMaintenanceMode()` manually in each handler/event — easy to forget. Go: one gate in the router.
- Owner ID typed and validated at startup (fatal if `/maintenance` wired but env missing → log warning, disable command).

## Acceptance

- Toggle works, gate blocks non-owner interactions, crons skip, owner unaffected.
