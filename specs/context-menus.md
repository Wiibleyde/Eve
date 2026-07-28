# Spec: Context Menus

## Goal

Three context menu commands (right-click actions).

Depends on: [00-interactions-router.md](00-interactions-router.md) (context menu dispatch).

## 1. Message → «Créer une citation»

- Right-click a message → creates a quote from it.
- Author = message author, text = message content, context = empty.
- Reuses the existing quote feature pipeline (`internal/bot/features/quote/`): same image generation (`image.go`), same posting to the configured `quote.channel`, same DB insert.
- Empty content (embed-only/attachment-only message) → ephemeral error «Ce message ne contient pas de texte.»
- Implementation: extract quote-creation into a shared function `quote.Create(ctx, guildID, authorID, text, context)` used by both the slash command and the context menu.

## 2. User → «Récupèrer la photo de profil»

- Ephemeral embed, title «Photo de profil», image = user display avatar PNG 1024. Prefer **guild-specific avatar** when present (TS used global only).

## 3. User → «Récupèrer la bannière»

- Ephemeral embed with the user's banner image 1024.
- Requires a REST fetch of the full user object (banner not in cache/partial user).
- No banner → ephemeral «Cet utilisateur n'a pas de bannière.»

## Registration

Each lives in its owning feature (quote ctx menu in `features/quote/`, avatar+banner in a small `features/userinfo/`). Commands appended to `allCommands` like slash commands.

## Acceptance

- All three appear in right-click menus and behave as described; quote ctx menu output identical to `/quote add`.
