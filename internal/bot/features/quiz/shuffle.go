package quiz

import (
	"errors"
	"math/rand/v2"
	"strconv"
	"strings"
)

const answerCount = 4

const goodAnswerIndex = 0

type permutation [answerCount]int

var errBadPermutation = errors.New("quiz: invalid shuffle permutation")

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

func encodePermutation(perm permutation) string {
	parts := make([]string, 0, answerCount)
	for _, v := range perm {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

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

func applyPermutation(answers [answerCount]string, perm permutation) [answerCount]string {
	var out [answerCount]string
	for i, src := range perm {
		out[i] = answers[src]
	}
	return out
}
