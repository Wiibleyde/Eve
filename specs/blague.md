# Spec: Blague (jokes)

## Goal

Random French joke from [blagues-api.fr](https://www.blagues-api.fr/), answer hidden behind a spoiler, with a "make public" button.

Depends on: [00-interactions-router.md](00-interactions-router.md).

## Command

`/blague` — option `type`* with choices:

| Label | API category |
|---|---|
| Générale | `global` |
| Développeur | `dev` |
| Beauf | `beauf` |

(`dark`, `limit`, `blondes` intentionally excluded, same as TS.)

## Flow

1. Ephemeral reply: embed with the joke as description, field «Réponse :» = `||answer||`, footer disclaimer «⚠️ Eve et ses développeurs ne sont pas responsables des blagues proposées. ⚠️».
2. Button «Rendre publique» — custom ID `blague:public`.
3. On click: post the same embed publicly in the channel (attributed «Demandée par @user»), remove the button from / delete the ephemeral.

Since ephemeral messages can't be reliably fetched later, encode what's needed in the flow: either post directly from the button interaction's message embed (it's available on the interaction), or store joke ID in the custom ID (`blague:public:<jokeID>`) and re-fetch from the API by ID. **Prefer reading the embed off the interaction message** — zero state.

## API client

No Go SDK needed — plain REST:

```
GET https://www.blagues-api.fr/api/type/<category>/random
Authorization: Bearer <BLAGUE_API_TOKEN>
→ {"id":..,"type":"..","joke":"..","answer":".."}
```

Small client in the feature package, 5s timeout, API error → ephemeral French error embed.

## Env

```
BLAGUE_API_TOKEN=...   # optional; if unset, command replies "feature désactivée"
```

Feature registers its command only when the token is present (avoid dead command).

## Improvements vs TS

- No npm-style SDK dependency; direct REST.
- Command not registered at all when unconfigured (TS registered it and failed at runtime).
- Public button works stateless from the interaction message.

## Acceptance

- Each category returns jokes; button republishes publicly; missing token hides command.
