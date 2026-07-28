# Spec: REST API completion

## Goal

Bring the Fiber API (`internal/api/`) to parity with the TS express API, minus LSMS.

## Current state

Only `GET /api/v1/status`.

## Endpoints

All under `/api/v1`, JSON:

| Route | Response |
|---|---|
| `GET /status` | exists — keep |
| `GET /health` | `{"health":"good","uptime":<seconds float>}` |
| `GET /info` | `{"app":"Eve - API","version":"<build version>","author":"Wiibleyde"}` — version injected at build (`-ldflags -X`), not hardcoded `0.1.0` like TS |
| `GET /ping` | `{"message":"pong","timestamp":"<RFC3339>"}` |
| `GET /loto/stats` | active loto summary: game name, tickets sold, pot, players count — see [loto.md](loto.md) |
| `GET /loto/winners` | finished games: prizes with winner names + ticket numbers |

`/status` vs `/health` vs `/ping` are near-duplicates in TS; keep all three for compatibility (cheap), but implement via one handler file.

## CORS

Add Fiber CORS middleware: origin `*`, methods GET only (API is read-only — TS allowed POST/PUT/DELETE with zero write routes; don't), no credentials.

## Structure

- `internal/api/controllers/` one file per domain (`status.go` exists; add `loto.go`).
- Controllers read via the ent client / feature packages — no SQL in controllers.

## Improvements vs TS

- Read-only CORS surface.
- Build-time version.
- Single source for status-ish endpoints.

## Acceptance

- All routes return documented shapes; CORS preflight passes for GET.
