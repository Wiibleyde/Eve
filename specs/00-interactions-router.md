# Spec: Interactions Router Extension

## Goal

Extend `internal/bot/router/router.go` beyond slash commands to dispatch **buttons, modals, select menus, and context menus**. This is the foundation for quiz, motus, blague, loto, and context-menu features.

## Current state

Router only supports `OnCommand(name, fn)` for `ApplicationCommandInteractionCreate` (slash). Everything else is dropped.

## Custom ID scheme

Single formalized scheme (TS used ad-hoc `name--data`):

```
<feature>:<action>[:<data>...]
```

- Separator: `:` (data segments must not contain it; use base64/UUIDs if needed).
- Max 100 chars (Discord limit) — keep data to UUIDs/ints.
- Examples: `quiz:answer:0:<questionID>`, `loto:buy:<gameUUID>`, `motus:try`.

Router matches on the **`<feature>:<action>` prefix**, passes remaining segments to the handler as `[]string`.

## API

```go
type Router struct { ... }

// existing
func (r *Router) OnCommand(name string, fn CommandHandler)

// new
func (r *Router) OnButton(prefix string, fn ButtonHandler)         // prefix = "feature:action"
func (r *Router) OnModal(prefix string, fn ModalHandler)
func (r *Router) OnSelectMenu(prefix string, fn SelectMenuHandler)
func (r *Router) OnUserContextMenu(name string, fn UserCtxHandler)
func (r *Router) OnMessageContextMenu(name string, fn MessageCtxHandler)
```

Handler signatures mirror disgo event types:

```go
type ButtonHandler func(e *events.ComponentInteractionCreate, args []string)
type ModalHandler func(e *events.ModalSubmitInteractionCreate, args []string)
type SelectMenuHandler func(e *events.ComponentInteractionCreate, args []string)
type UserCtxHandler func(e *events.ApplicationCommandInteractionCreate)      // data.Type == User
type MessageCtxHandler func(e *events.ApplicationCommandInteractionCreate)   // data.Type == Message
```

`Attach(client)` registers listeners for `ComponentInteractionCreate` and `ModalSubmitInteractionCreate` in addition to the existing command listener. Context menus dispatch inside the existing `ApplicationCommandInteractionCreate` listener by command type.

## Dispatch rules

- Parse custom ID by splitting on `:`; look up `parts[0] + ":" + parts[1]` in the relevant map; pass `parts[2:]` as args.
- Unknown custom ID → log at DEBUG, respond ephemeral error ("Interaction inconnue ou expirée.") so the user isn't left with a spinner.
- Handler panics: recover in the dispatch layer, log ERROR, respond generic ephemeral error. (One broken feature must not crash the bot.)

## Registration of context menu commands

Context menu commands are `discord.ApplicationCommandCreate` entries (type User/Message) — append to `allCommands` in `bot.go` exactly like slash commands.

## Improvements vs TS

- TS: four separate registries (`buttons.ts`, `modals.ts`, `selectMenus.ts`, contextMenu files) with per-file maps and inconsistent `--` separators. Go: one router, one scheme.
- TS: unknown button IDs silently ignored → eternal "thinking" state. Go: explicit expired-interaction reply.
- Panic isolation per handler (TS relied on top-level try/catch in interactionCreate).

## Acceptance

- Ping-style test feature can register one button + one modal and round-trip data through custom ID args.
- Unknown/expired custom ID gets an ephemeral French error.
