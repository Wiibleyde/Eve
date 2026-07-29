package loto

import (
	"fmt"
	"sort"
	"strings"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/router"
	"Eve/internal/database/ent"

	"github.com/disgoorg/disgo/discord"
)

const (
	// Discord hard limits.
	maxEmbedDescription = 4096
	maxFieldValue       = 1024
	// Only the busiest players are listed to keep the embed readable.
	maxListedPlayers = 50
)

// tally is a player with its ticket count, used for the leaderboard.
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

// leaderboard sorts players by ticket count desc, then name asc.
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

// buildEmbed renders the public loto embed: sales summary, player leaderboard
// and the ranked prize list (with winners once the draw happened).
func buildEmbed(s *snapshot) discord.Embed {
	game := s.game
	embed := embeds.BaseEmbed()
	embed.Title = fmt.Sprintf("🎟️ Loto: %s 🎟️", game.Name)

	ticketsSold := len(s.tickets)
	lines := []string{
		"⚠️ __**Attention, l'écriture des noms est sensible à la case !**__",
		fmt.Sprintf("Nombre de tickets vendus: **%d**", ticketsSold),
	}

	if game.CooldownMinutes > 0 {
		lines = append(lines, fmt.Sprintf("Cooldown par joueur: **%d %s**", game.CooldownMinutes, pluralize(game.CooldownMinutes, "minute", "minutes")))
		if game.MaxTicketsPerPurchase > 0 {
			lines = append(lines, fmt.Sprintf("Limite par achat: **%d ticket(s)**", game.MaxTicketsPerPurchase))
		}
	} else if game.MaxTicketsPerPurchase > 0 {
		lines = append(lines, fmt.Sprintf("Limite par achat: **%d ticket(s)**", game.MaxTicketsPerPurchase))
	}

	if len(s.prizes) > 0 {
		lines = append(lines, fmt.Sprintf("Nombre de gains: **%d**", len(s.prizes)))
	}
	if ticketsSold > 0 {
		lines = append(lines, fmt.Sprintf("Cagnotte actuelle: **%d**$", s.pot()))
	}
	if !game.Active {
		lines = append(lines, "", "🔒 **Ce loto est terminé, le tirage a été effectué.**")
	}

	board := s.leaderboard()
	if len(board) == 0 {
		lines = append(lines, "", "Aucun ticket vendu pour le moment.")
	} else {
		if len(board) > maxListedPlayers {
			board = board[:maxListedPlayers]
		}
		lines = append(lines, "")
		for _, entry := range board {
			lines = append(lines, fmt.Sprintf("- %s (%d)", entry.name, entry.count))
		}
	}

	embed.Description = truncate(strings.Join(lines, "\n"), maxEmbedDescription)

	footerParts := []string{fmt.Sprintf("Prix du ticket: %d$", game.TicketPrice)}
	if game.CooldownMinutes > 0 {
		footerParts = append(footerParts, fmt.Sprintf("Cooldown: %d min", game.CooldownMinutes))
	}
	if game.MaxTicketsPerPurchase > 0 {
		footerParts = append(footerParts, fmt.Sprintf("Limite achat: %d", game.MaxTicketsPerPurchase))
	}
	icon := ""
	if embed.Footer != nil {
		icon = embed.Footer.IconURL
	}
	embed.Footer = &discord.EmbedFooter{Text: strings.Join(footerParts, " | "), IconURL: icon}

	if len(s.prizes) > 0 {
		chunks := chunkLines(prizeLines(s), maxFieldValue)
		for i, chunk := range chunks {
			name := "🎁 Gains"
			if i > 0 {
				name = "🎁 Gains (suite)"
			}
			embed.Fields = append(embed.Fields, discord.EmbedField{Name: name, Value: chunk})
		}
	}

	return embed
}

// prizeLines renders one line per prize, in position order.
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

// sellerSummary lists how much each seller collected, for the closing embed.
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
		lines = append(lines, fmt.Sprintf("<@%s> : %d ticket(s) - **%d$**", sellerID, counts[sellerID], counts[sellerID]*ticketPrice))
	}
	return truncate(strings.Join(lines, "\n"), maxFieldValue)
}

// buildComponents returns the management button row of an active game.
// A finished game has no components at all.
func buildComponents(gameID string) []discord.LayoutComponent {
	return []discord.LayoutComponent{
		discord.NewActionRow(
			discord.NewPrimaryButton("Acheter", router.BuildCustomID(buttonBuy, gameID)).
				WithEmoji(discord.NewComponentEmoji("🎟️")),
			discord.NewSecondaryButton("Retirer des tickets", router.BuildCustomID(buttonRemove, gameID)).
				WithEmoji(discord.NewComponentEmoji("🗑️")),
			discord.NewSecondaryButton("Corriger un nom", router.BuildCustomID(buttonEditPlayer, gameID)).
				WithEmoji(discord.NewComponentEmoji("✏️")),
			discord.NewDangerButton("Tirer au sort", router.BuildCustomID(buttonDraw, gameID)).
				WithEmoji(discord.NewComponentEmoji("🎲")),
		),
	}
}

// chunkLines packs lines into chunks of at most maxLength bytes — Discord's
// field limit is in characters, so byte-counting is the conservative side.
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

// truncate cuts a string down to max bytes (ellipsis included) without ever
// splitting a UTF-8 rune, which Discord rejects.
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

// utf8Start reports whether b starts a UTF-8 rune.
func utf8Start(b byte) bool { return b&0xC0 != 0x80 }

func pluralize(n int, singular, plural string) string {
	if n <= 1 {
		return singular
	}
	return plural
}
