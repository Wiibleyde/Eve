# Spec: Loto

## Goal

RP lottery ("SABS loto"): named games with ticket sales (by RP character name), configurable ticket price, per-player purchase cooldown, ranked prizes, and a public draw. Managed through a persistent public message with buttons.

Depends on: [00-interactions-router.md](00-interactions-router.md).

## Commands

`/loto` (guild only):

- `create` — options: `name`* (≤50), `prize1`* … up to `prize10` (labels, at least 1), `ticketprice` (int, default 500), `cooldown` (minutes between purchases per player, default 0), `maxtickets` (max tickets per purchase; **required if cooldown > 0**).
  - Only one active game per guild — creating while one is active → ephemeral error.
  - Posts the public loto embed with buttons.

Management is button-driven from the message (no other subcommands needed).

## Data model (ent)

- `loto_game`: `id`, `guild_id`, `name`, `active` (bool default true), `ticket_price` (default 500), `cooldown_minutes` (default 0), `max_tickets_per_purchase` (nullable), `message_id`, `channel_id`, `created_at`.
- `loto_player`: `id`, `game_id` (FK cascade), `name` (RP name, **case-sensitive**), `last_play` (for cooldown). Unique (`game_id`, `name`).
- `loto_ticket`: `id`, `game_id` (FK cascade), `player_id` (FK cascade), `seller_id` (Discord snowflake of who sold), `number` (int, sequential per game — TS had no ticket number column yet referenced `winningTicketNumber`; make it explicit).
- `loto_prize`: `id`, `game_id` (FK cascade), `label`, `position` (unique per game), `winner_player_id` (nullable), `winning_ticket_number` (nullable), `drawn_at` (nullable).

## Public message

Embed (see TS `generateLotoEmbed` for layout):

- Title «🎟️ Loto: <name> 🎟️», warning that names are case-sensitive.
- Tickets sold, pot (`sold × price`), cooldown/limit info, top-50 players by ticket count, prizes list with winners once drawn.
- Footer: ticket price + cooldown + limit.

Buttons:

- «Acheter» `loto:buy:<gameID>` — anyone.
- «Retirer des tickets» `loto:remove:<gameID>` — seller/admin.
- «Corriger un nom» `loto:editplayer:<gameID>` — admin.
- «Tirer au sort» `loto:draw:<gameID>` — admin.

Admin = a Discord permission check (TS used `PinMessages` as proxy; keep **Manage Messages** — clearer).

## Flows

### Buy `loto:buy` → modal `loto:buymodal:<gameID>`
Inputs: `playerName` (≤50), `count` (int ≥1). On submit:
1. Game must be active.
2. Upsert player by exact name.
3. Cooldown: if `cooldown_minutes > 0` and `now - last_play < cooldown` → ephemeral error with remaining time. Enforce `count ≤ max_tickets_per_purchase`.
4. Create `count` tickets (sequential `number`), update `last_play`, edit public embed, ephemeral receipt (total cost).

### Remove tickets `loto:remove` → modal
Inputs: player name, count. Delete newest N tickets of that player; player with 0 tickets stays (history). Edit embed.

### Edit player name `loto:editplayer` → modal
Inputs: old name, new name. Rename (unique conflict → error). Edit embed.

### Draw `loto:draw`
1. Admin check; game active; at least 1 ticket; prizes exist.
2. For each prize in `position` order: pick a uniformly random ticket among tickets **not already winning** (one win per ticket; a player may win multiple prizes via different tickets — same as TS logic). Record winner + ticket number + `drawn_at`.
3. Set game `active = false`, edit embed with winners, post public announcement message listing prizes → winners.
4. Draw is **one-shot and final** — confirmation button first («Confirmer le tirage» `loto:drawconfirm:<gameID>`) since it's irreversible.

## API exposure

`GET /api/v1/loto/stats`, `GET /api/v1/loto/winners` — see [api.md](api.md).

## Improvements vs TS

- Explicit sequential ticket numbers.
- One active game per guild enforced.
- Draw confirmation step.
- Manage Messages instead of PinMessages proxy.

## Acceptance

- Create → buy (cooldown + limit enforced) → draw assigns all prizes → embed + announcement correct; API endpoints reflect the game.
