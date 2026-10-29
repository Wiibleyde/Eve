package loto

import (
	"fmt"
	"strings"
	"testing"

	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"
)

func crowdedSnapshot(active bool) *snapshot {
	s := &snapshot{game: &ent.LotoGame{
		ID:                    "game",
		Name:                  strings.Repeat("N", maxGameNameLength),
		Active:                active,
		TicketPrice:           500,
		CooldownMinutes:       maxCooldownMinutes,
		MaxTicketsPerPurchase: maxTicketsPerBuy,
	}}
	for i := range 200 {
		id := fmt.Sprintf("player%d", i)
		s.players = append(s.players, &ent.LotoPlayer{ID: id, Name: fmt.Sprintf("%03d%s", i, strings.Repeat("é", 23))})
		s.tickets = append(s.tickets, &ent.LotoTicket{PlayerID: id, SellerID: fmt.Sprintf("1%018d", i), Number: i + 1})
	}
	for i := range maxPrizeCount {
		prize := &ent.LotoPrize{ID: fmt.Sprintf("prize%d", i), Label: strings.Repeat("x", maxPrizeLabelLen), Position: i}
		if !active {
			prize.WinnerPlayerID = fmt.Sprintf("player%d", i)
			prize.WinningTicketNumber = i + 1
		}
		s.prizes = append(s.prizes, prize)
	}
	return s
}

func TestCardsFitDiscordTextLimit(t *testing.T) {
	for _, active := range []bool{true, false} {
		s := crowdedSnapshot(active)
		cards := map[string]*ui.Card{
			"public":  buildCard(s).Subtext("Nouveau loto créé par <@123456789012345678>"),
			"menu":    prizeMenuCard(s, "prize0", strings.Repeat("n", 300)),
			"compact": buildCard(&snapshot{game: s.game, players: s.players[:3], tickets: s.tickets[:3], prizes: s.prizes[:1]}),
		}
		for name, card := range cards {
			got := card.TextLength()
			if got > ui.MaxText {
				t.Errorf("active=%v %s card: %d characters, limit %d", active, name, got, ui.MaxText)
			}
			if name != "compact" && got < ui.MaxText/2 {
				t.Errorf("active=%v %s card: only %d characters measured, budget left unused", active, name, got)
			}
		}
	}
}

func TestCompactCardKeepsEveryLine(t *testing.T) {
	s := crowdedSnapshot(true)
	s.players, s.tickets, s.prizes = s.players[:3], s.tickets[:3], s.prizes[:2]
	text := strings.Join(ui.Texts(buildCard(s).Components()), "\n")
	if strings.Contains(text, "… et") {
		t.Fatalf("small card was truncated:\n%s", text)
	}
}

func TestFitLines(t *testing.T) {
	lines := []string{"aaaa", "bbbb", "cccc"}
	more := func(hidden int) string { return fmt.Sprintf("+%d", hidden) }

	if got := fitLines(lines, 3, 100, more); len(got) != 3 {
		t.Errorf("everything fits: got %q", got)
	}
	if got := fitLines(lines, 3, 13, more); strings.Join(got, ",") != "aaaa,bbbb,+1" {
		t.Errorf("one hidden: got %q", got)
	}
	if got := fitLines(lines, 10, 100, more); strings.Join(got, ",") != "aaaa,bbbb,cccc,+7" {
		t.Errorf("capped list: got %q", got)
	}
	if got := fitLines(lines, 3, 1, more); got != nil {
		t.Errorf("no room: got %q", got)
	}
}

func TestAnnounceChunksKeepEveryWinner(t *testing.T) {
	lines := winnerLines(crowdedSnapshot(false))
	chunks := chunkLines(lines, announceChunkLength)
	if got := strings.Count(strings.Join(chunks, "\n"), "→"); got != maxPrizeCount {
		t.Fatalf("announced %d winners, want %d", got, maxPrizeCount)
	}
	for _, chunk := range chunks {
		if len(chunk) > announceChunkLength {
			t.Errorf("chunk of %d bytes exceeds %d", len(chunk), announceChunkLength)
		}
	}
}
