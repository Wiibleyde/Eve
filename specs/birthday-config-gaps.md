# Spec: Gaps in existing features (birthday, config)

Small completions of features already ported.

## Birthday

Missing subcommand vs TS:

- `/birthday get` — ephemeral: «Votre date d'anniversaire est le <formatted> (<t:ts:R>)», or «Vous n'avez pas enregistré votre anniversaire.» if none.

Everything else (`set`, `remove`, `list`, `adminset`, scheduler) exists — no change.

## Config

Go currently hardcodes two subcommands (`birthday-channel`, `quote-channel`). TS had generic `set/get/reset/list` over option keys.

**Decision: keep the Go typed-subcommand approach** (it gives per-key option types + validation — the TS generic string-key design was the bad practice). But make adding keys cheap:

1. Internal registry in the config feature:

```go
type Key struct {
    Name        string   // e.g. "birthday.channel"
    Command     string   // subcommand name, e.g. "birthday-channel"
    Description string
    Kind        Kind     // Channel | Role | Bool | String
}
var Keys = []Key{...}
```

Subcommands (`<key>` setter, plus shared `get`, `reset`, `list`) generated from the registry.

2. New keys needed by the specs in this directory (register as their features land):

| Key | Kind | Used by |
|---|---|---|
| `birthday.channel` | Channel | exists |
| `quote.channel` | Channel | exists |
| `calendar.url` / `calendar.channel` / `calendar.message` | internal (not user-settable via /config; managed by `/calendar`) | calendar |
| `jokes.disabled` | Bool | message-reactions |
| `debug.role` | internal | debug |

3. `/config list` shows current values of all **user-visible** keys; `get`/`reset` take a key choice built from the registry.

Permission: Manage Guild (admin), unchanged.

## Acceptance

- `/birthday get` works.
- Adding a config key = one registry entry; `list/get/reset` pick it up automatically.
