package quiz

import "testing"

func TestShufflePermutationIsAPermutation(t *testing.T) {
	for i := 0; i < 200; i++ {
		perm := shufflePermutation()
		var seen [answerCount]bool
		for _, v := range perm {
			if v < 0 || v >= answerCount || seen[v] {
				t.Fatalf("not a permutation: %v", perm)
			}
			seen[v] = true
		}
	}
}

func TestShufflePermutationCoversEveryPosition(t *testing.T) {
	var positions [answerCount]int
	for i := 0; i < 500; i++ {
		perm := shufflePermutation()
		for pos, src := range perm {
			if src == goodAnswerIndex {
				positions[pos]++
			}
		}
	}
	for pos, count := range positions {
		if count == 0 {
			t.Fatalf("good answer never shown on button %d", pos)
		}
	}
}

func TestEncodeDecodePermutationRoundTrip(t *testing.T) {
	perm := permutation{2, 0, 3, 1}
	encoded := encodePermutation(perm)
	if encoded != "2,0,3,1" {
		t.Fatalf("encodePermutation = %q, want %q", encoded, "2,0,3,1")
	}
	decoded, err := decodePermutation(encoded)
	if err != nil {
		t.Fatalf("decodePermutation: %v", err)
	}
	if decoded != perm {
		t.Fatalf("decodePermutation = %v, want %v", decoded, perm)
	}
}

func TestDecodePermutationRejectsInvalid(t *testing.T) {
	for _, in := range []string{"", "0,1,2", "0,1,2,3,4", "0,1,2,9", "0,1,1,2", "a,b,c,d", "-1,1,2,3"} {
		if _, err := decodePermutation(in); err == nil {
			t.Fatalf("decodePermutation(%q) should have failed", in)
		}
	}
}

func TestApplyPermutation(t *testing.T) {
	answers := [answerCount]string{"good", "bad1", "bad2", "bad3"}
	got := applyPermutation(answers, permutation{2, 0, 3, 1})
	want := [answerCount]string{"bad2", "good", "bad3", "bad1"}
	if got != want {
		t.Fatalf("applyPermutation = %v, want %v", got, want)
	}
}
