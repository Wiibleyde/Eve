package quiz

import (
	"errors"
	"math/rand/v2"
	"strconv"
	"strings"
)

// answerCount is the fixed number of answers of a question: one good, three bad.
const answerCount = 4

// goodAnswerIndex is the position of the good answer in storage order
// [good, bad1, bad2, bad3]. The shuffle maps displayed positions to it.
const goodAnswerIndex = 0

// permutation maps a displayed button position to a stored answer index:
// perm[i] is the stored index of the answer shown on button i.
type permutation [answerCount]int

var errBadPermutation = errors.New("quiz: invalid shuffle permutation")

// shufflePermutation returns a uniformly random permutation using Fisher-Yates.
//
// The TS version used sort(() => Math.random()-0.5), which is not a uniform
// shuffle: with 4 elements it visibly favours some positions, and the good
// answer ended up under the same button far too often.
func shufflePermutation() permutation {
	var perm permutation
	for i := range perm {
		perm[i] = i
	}
	for i := len(perm) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		perm[i], perm[j] = perm[j], perm[i]
	}
	return perm
}

// encodePermutation renders a permutation as "2,0,3,1" for storage.
func encodePermutation(perm permutation) string {
	parts := make([]string, 0, answerCount)
	for _, v := range perm {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

// decodePermutation parses a stored shuffle back, rejecting anything that is
// not an actual permutation of 0..3 (a corrupted row must not silently mark
// every answer wrong).
func decodePermutation(s string) (permutation, error) {
	var perm permutation
	parts := strings.Split(s, ",")
	if len(parts) != answerCount {
		return perm, errBadPermutation
	}
	var seen [answerCount]bool
	for i, part := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || v < 0 || v >= answerCount || seen[v] {
			return perm, errBadPermutation
		}
		seen[v] = true
		perm[i] = v
	}
	return perm, nil
}

// applyPermutation reorders the stored answers into display order.
func applyPermutation(answers [answerCount]string, perm permutation) [answerCount]string {
	var out [answerCount]string
	for i, src := range perm {
		out[i] = answers[src]
	}
	return out
}
