// Package motus implements the Motus / Wordle word game.
//
// A `/motus` call posts a public board message carrying an «Essayer» button.
// Anyone in the channel may press it: the button opens a modal, the modal
// submission is scored with the classic two-pass algorithm, the public board is
// edited and the player gets an ephemeral feedback. Six failed attempts lose
// the game, a full row of found letters wins it.
//
// Differences with the TypeScript version:
//
//   - games live in the active_motus table instead of an in-memory map, so a
//     restart leaves running boards playable;
//   - the external word API (trouve-mot.fr) is only a preference: a bundled
//     word list (assets/motus/words.txt) takes over when it is down;
//   - the answer is never logged above DEBUG (the TS version logged it at INFO,
//     which leaked it to anyone reading the console).
//
// Custom IDs follow the router scheme:
//
//	motus:try                  button on the board message
//	motus:submit:<messageID>   modal opened by that button
package motus

import "time"

const (
	// CustomIDTry is the board button prefix.
	CustomIDTry = "motus:try"
	// CustomIDSubmit is the guess modal prefix; the board message ID is
	// appended as the single data segment.
	CustomIDSubmit = "motus:submit"
	InputGuess     = "guess"

	CommandName = "motus"
)

// GameTTL is how long a board stays playable. Expiry is evaluated lazily, on
// the next interaction touching the game — no sweeper goroutine.
const GameTTL = 24 * time.Hour
