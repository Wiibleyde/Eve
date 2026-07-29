package loto

import (
	"fmt"
	"sort"
	"strings"

	"Eve/internal/bot/router"
	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"

	"github.com/disgoorg/disgo/discord"
)

const (
	maxBodyLength    = 2000
	maxSectionLength = 900
	maxListedPlayers = 50
)

const lotoColor = 0x9B59B6

type tally struct {
	name  string
	count int
}

func (s *snapshot) playerNames() map[string]string {
	names := make(map[string]string, len(s.players))
	for _, p := range s.players {
		names[p.ID] = p.Name
	}
	return names
}

func (s *snapshot) leaderboard() []tally {
	names := s.playerNames()
	counts := make(map[string]int, len(s.players))
	for _, t := range s.tickets {
		counts[t.PlayerID]++
	}

	list := make([]tally, 0, len(counts))
	for playerID, count := range counts {
		name := names[playerID]
		if name == "" {
			name = "Nom inconnu"
		}
		list = append(list, tally{name: name, count: count})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].count != list[j].count {
			return list[i].count > list[j].count
		}
		return list[i].name < list[j].name
	})
	return list
}

func (s *snapshot) pot() int {
	return len(s.tickets) * s.game.TicketPrice
}

func buildCard(s *snapshot) *ui.Card {
	game := s.game

	card := ui.New().
		Accent(lotoColor).
		Titlef("🎟️ Loto : %s", game.Name).
		Text("⚠️ __**Attention, l'écriture des noms est sensible à la case !**__")

	stats := []ui.Field{
		{Name: "Tickets vendus", Value: fmt.Sprintf("**%d**", len(s.tickets)), Inline: true},
	}
	if game.CooldownMinutes > 0 {
		stats = append(stats, ui.Field{
			Name:   "Cooldown par joueur",
			Value:  fmt.Sprintf("**%d %s**", game.CooldownMinutes, pluralize(game.CooldownMinutes, "minute", "minutes")),
			Inline: true,
		})
	}
	if game.MaxTicketsPerPurchase > 0 {
		stats = append(stats, ui.Field{
			Name:   "Limite par achat",
			Value:  fmt.Sprintf("**%d ticket(s)**", game.MaxTicketsPerPurchase),
			Inline: true,
		})
	}
	if len(s.prizes) > 0 {
		stats = append(stats, ui.Field{Name: "Nombre de gains", Value: fmt.Sprintf("**%d**", len(s.prizes)), Inline: true})
	}
	if len(s.tickets) > 0 {
		stats = append(stats, ui.Field{Name: "Cagnotte actuelle", Value: fmt.Sprintf("**%d$**", s.pot()), Inline: true})
	}
	card.Fields(stats...)

	if !game.Active {
		card.Text("🔒 **Ce loto est terminé, le tirage a été effectué.**")
	}

	card.Divider().Heading("👥 Participants")
	board := s.leaderboard()
	if len(board) == 0 {
		card.Text("Aucun ticket vendu pour le moment.")
	} else {
		if len(board) > maxListedPlayers {
			board = board[:maxListedPlayers]
		}
		lines := make([]string, 0, len(board))
		for _, entry := range board {
			lines = append(lines, fmt.Sprintf("- %s (%d)", entry.name, entry.count))
		}
		card.Text(truncate(strings.Join(lines, "\n"), maxBodyLength))
	}

	if len(s.prizes) > 0 {
		card.Divider().Heading("🎁 Gains")
		for _, chunk := range chunkLines(prizeLines(s), maxSectionLength) {
			card.Text(chunk)
		}
	}

	if !game.Active {
		card.Divider().Heading("📊 Ventes par vendeur").Text(sellerSummary(s.tickets, game.TicketPrice))
	} else {
		card.Row(actionButtons(game.ID)...)
	}

	footerParts := []string{fmt.Sprintf("Prix du ticket : %d$", game.TicketPrice)}
	if game.CooldownMinutes > 0 {
		footerParts = append(footerParts, fmt.Sprintf("Cooldown : %d min", game.CooldownMinutes))
	}
	if game.MaxTicketsPerPurchase > 0 {
		footerParts = append(footerParts, fmt.Sprintf("Limite achat : %d", game.MaxTicketsPerPurchase))
	}
	return card.Footer(strings.Join(footerParts, " · "))
}

func prizeLines(s *snapshot) []string {
	names := s.playerNames()
	lines := make([]string, 0, len(s.prizes))
	for i, prize := range s.prizes {
		rank := i + 1
		switch {
		case prize.WinnerPlayerID != "":
			winner := names[prize.WinnerPlayerID]
			if winner == "" {
				winner = "Nom inconnu"
			}
			ticketInfo := ""
			if prize.WinningTicketNumber > 0 {
				ticketInfo = fmt.Sprintf(" (ticket n°%d)", prize.WinningTicketNumber)
			}
			lines = append(lines, fmt.Sprintf("- **#%d** %s — 🎉 %s%s", rank, prize.Label, winner, ticketInfo))
		case s.game.Active:
			lines = append(lines, fmt.Sprintf("- **#%d** %s", rank, prize.Label))
		default:
			lines = append(lines, fmt.Sprintf("- **#%d** %s — Non attribué", rank, prize.Label))
		}
	}
	return lines
}

func sellerSummary(tickets []*ent.LotoTicket, ticketPrice int) string {
	if len(tickets) == 0 {
		return "Aucune vente enregistrée."
	}

	counts := make(map[string]int)
	order := make([]string, 0)
	for _, t := range tickets {
		if _, seen := counts[t.SellerID]; !seen {
			order = append(order, t.SellerID)
		}
		counts[t.SellerID]++
	}
	sort.SliceStable(order, func(i, j int) bool {
		if counts[order[i]] != counts[order[j]] {
			return counts[order[i]] > counts[order[j]]
		}
		return order[i] < order[j]
	})

	lines := make([]string, 0, len(order))
	for _, sellerID := range order {
		lines = append(lines, fmt.Sprintf("<@%s> : %d ticket(s) — **%d$**", sellerID, counts[sellerID], counts[sellerID]*ticketPrice))
	}
	return truncate(strings.Join(lines, "\n"), maxSectionLength)
}

func actionButtons(gameID string) []discord.InteractiveComponent {
	return []discord.InteractiveComponent{
		discord.NewPrimaryButton("Acheter", router.BuildCustomID(buttonBuy, gameID)).
			WithEmoji(discord.NewComponentEmoji("🎟️")),
		discord.NewSecondaryButton("Retirer des tickets", router.BuildCustomID(buttonRemove, gameID)).
			WithEmoji(discord.NewComponentEmoji("🗑️")),
		discord.NewSecondaryButton("Corriger un nom", router.BuildCustomID(buttonEditPlayer, gameID)).
			WithEmoji(discord.NewComponentEmoji("✏️")),
		discord.NewDangerButton("Tirer au sort", router.BuildCustomID(buttonDraw, gameID)).
			WithEmoji(discord.NewComponentEmoji("🎲")),
	}
}

func chunkLines(lines []string, maxLength int) []string {
	chunks := make([]string, 0, 1)
	current := ""

	for _, line := range lines {
		candidate := line
		if current != "" {
			candidate = current + "\n" + line
		}
		if len(candidate) <= maxLength {
			current = candidate
			continue
		}
		if current == "" {
			chunks = append(chunks, truncate(line, maxLength))
			continue
		}
		chunks = append(chunks, current)
		current = truncate(line, maxLength)
	}

	if current != "" {
		chunks = append(chunks, current)
	}
	return chunks
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	const ellipsis = "…"
	if max <= len(ellipsis) {
		cut := max
		for cut > 0 && !utf8Start(s[cut]) {
			cut--
		}
		return s[:cut]
	}
	cut := max - len(ellipsis)
	for cut > 0 && !utf8Start(s[cut]) {
		cut--
	}
	return s[:cut] + ellipsis
}

func utf8Start(b byte) bool { return b&0xC0 != 0x80 }

func pluralize(n int, singular, plural string) string {
	if n <= 1 {
		return singular
	}
	return plural
}
