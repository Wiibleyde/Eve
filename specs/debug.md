# Spec: Debug role toggle

## Goal

`/debug` toggles a per-guild "debug" role on the invoking user. The role marks testers (e.g. to see debug output or bypass gates).

## Command

`/debug` (guild only). Flow:

1. Read `debug.role` from `guild_configs`; if missing or role deleted, create a role named `Eve Debug` (no permissions, not hoisted, not mentionable) and store its ID.
2. User has role → remove → «Vous n'êtes plus en mode debug sur le serveur <name>».
   User lacks role → add → «Vous êtes maintenant en mode debug sur le serveur <name>».
3. Ephemeral replies. Missing `Manage Roles` bot permission → ephemeral error explaining it.

## Access control

TS let **anyone** self-assign — keep that (it's a harmless marker role), but the created role must always be permissionless. Verify on each use that the stored role still has zero permissions; if someone escalated it, refuse and warn.

## Data model

`guild_configs` key `debug.role` = role snowflake. (TS used a `GuildData.debugRoleId` column — key/value store is enough.)

## Improvements vs TS

- Role stored in `guild_configs`, no bespoke table.
- Permission-escalation check on the managed role.
- Handles deleted-role recreation cleanly.

## Acceptance

- Toggle on/off works; role auto-created once; deleting the role manually self-heals.
