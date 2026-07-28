# Spec: Quiz

## Goal

Community quiz: launch a random question with 4 shuffled answer buttons, track per-user right/wrong stats, leaderboard.

Depends on: [00-interactions-router.md](00-interactions-router.md).

## Commands

`/quiz` (guild only):

- `launch` — post a random question publicly in the channel.
- `create` — options: `question`* `answer`* `bad1`* `bad2`* `bad3`* `category`* `difficulty`* (choices: `facile`/`normal`/`difficile`). Ephemeral confirmation.
- `leaderboard` — option `choice`* (choices: `best_scores` «Meilleurs scores», `best_ratios` «Meilleurs ratios», `worst_scores` «Scores les plus bas»). Public embed, top 10.

## Data model (ent)

Existing tables `quiz`, `quiz_question`, `quiz_answer`, `quiz_user_answer` already exist on this branch — **reconcile with this design** (adjust schemas if they diverge):

- `quiz_question`: `id`, `question` (unique, ≤2048), `good_answer`, `bad_answer_1..3`, `category`, `difficulty`, `author_id` (snowflake, nullable), `guild_id`, `created_at`, `last_used_at` (nullable).
- `active_quiz` (replaces TS in-memory map — **must survive restarts**):
  - `id`, `question_id` (FK), `message_id` (unique), `channel_id`, `guild_id`, `shuffle` (string, e.g. `"2,0,3,1"` — permutation of answers), `launched_at`, `expires_at`.
- `quiz_stat`: `id`, `user_id` (snowflake, unique), `good_answers` (int, default 0), `bad_answers` (int, default 0).
  - TS put these on `GlobalUserData`; Go keeps a dedicated table (no god-table).
- `quiz_user_answer`: `id`, `active_quiz_id` (FK), `user_id`, `correct` (bool), `answered_at`. Unique on (`active_quiz_id`, `user_id`) — one answer per user per quiz.

## Flow

### Launch
1. Count questions; 0 → ephemeral error «Aucun quiz n'a été créé...».
2. Pick random question — prefer least-recently-used: `ORDER BY last_used_at ASC NULLS FIRST, random tiebreak` (TS pure random repeats questions; LRU is the fix). Update `last_used_at`.
3. Shuffle the 4 answers with **Fisher–Yates** (TS used `sort(() => Math.random()-0.5)` — biased).
4. Public embed: question in code block, category, difficulty, author mention, expiry timestamp (`<t:...:R>`). 4 buttons custom IDs `quiz:answer:<idx>` (idx 0–3 = shuffled position).
5. Insert `active_quiz` row with `expires_at = now + 8h`.
6. Ephemeral success to launcher.

### Answer button `quiz:answer:<idx>`
1. Look up `active_quiz` by `message_id` (**not** by parsing the embed text like TS did).
2. Missing row or expired → ephemeral: quiz expired + reveal correct answer (fetch via `question_id`; if row purged, fall back to question text lookup).
3. User already answered (unique constraint) → ephemeral «Vous avez déjà répondu...».
4. Map `idx` through `shuffle` → correct or not. Record `quiz_user_answer`, increment `quiz_stat` counters (upsert), ephemeral result embed (green/red, correct answer shown on wrong).

### Expiry
Birthday-style scheduler or lazy check on interaction (lazy is enough — button handler already validates `expires_at`). Optional daily cleanup deleting `active_quiz` rows older than 7 days.

### Leaderboard
- `best_scores`: order by `good_answers` desc.
- `best_ratios`: `good/(good+bad)` desc, minimum 10 total answers (avoid 1/1 = 100%).
- `worst_scores`: order by `bad_answers` desc.
Top 10, display username mentions + stats.

## Embeds

Quiz-branded author header. **Do not** use the TS hardcoded `cdn.discordapp.com/attachments/...` icon URL (signed, expires) — commit an icon under `assets/` or omit.

## Improvements vs TS (summary)

- Active games in DB, not RAM → survive restarts.
- Custom ID carries answer index; no embed-text parsing.
- Fisher–Yates shuffle; LRU question pick.
- One-answer-per-user enforced by unique constraint, not two in-memory arrays.
- Ratio leaderboard has a minimum-answers floor.

## Acceptance

- Launch → answer → stats increment; restart bot mid-quiz → buttons still work; expired quiz reveals answer.
