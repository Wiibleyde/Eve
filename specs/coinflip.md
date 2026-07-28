# Spec: Coinflip

## Goal

Flip a coin.

## Command

`/coinflip` — public reply, base embed: «Le résultat du lancer de pièce est : **Pile**.» or **Face**.

## Implementation

`crypto/rand`-seeded or `math/rand/v2` — one line. Feature package `internal/bot/features/coinflip/`.

Optional flourish (nice-to-have, not required): 1/6000 chance of «La pièce est tombée sur la tranche. 🪙».

## Acceptance

- Returns Pile/Face roughly 50/50.
