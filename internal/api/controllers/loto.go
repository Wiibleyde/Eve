package controllers

import (
	"context"
	"strconv"
	"time"

	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/lotogame"
	"Eve/internal/database/ent/lotoplayer"
	"Eve/internal/database/ent/lotoprize"
	"Eve/internal/database/ent/lototicket"
	"Eve/internal/logger"

	"github.com/gofiber/fiber/v3"
)

type LotoController struct{}

const lotoMaxWinnersLimit = 500

type LotoGameStats struct {
	ID           string    `json:"id"`
	GuildID      string    `json:"guildId" description:"Discord guild snowflake ID." example:"1091918939407908944"`
	Name         string    `json:"name"`
	Active       bool      `json:"active" description:"Whether the game is still running."`
	TicketPrice  int       `json:"ticketPrice" description:"Price of a single ticket."`
	TicketsSold  int       `json:"ticketsSold"`
	Pot          int       `json:"pot" description:"ticketsSold multiplied by ticketPrice."`
	PlayersCount int       `json:"playersCount"`
	PrizesCount  int       `json:"prizesCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

type LotoStatsTotals struct {
	Games       int `json:"games" description:"Number of games in the games array."`
	TicketsSold int `json:"ticketsSold"`
	Pot         int `json:"pot"`
	Players     int `json:"players" description:"Sum of per-game player counts. A player in two games counts twice."`
}

type LotoStatsResponse struct {
	Games  []LotoGameStats `json:"games" nullable:"false"`
	Totals LotoStatsTotals `json:"totals"`
}

type LotoWinner struct {
	GameID              string     `json:"gameId"`
	GameName            string     `json:"gameName" description:"Empty when the parent game could not be resolved."`
	GuildID             string     `json:"guildId" description:"Discord guild snowflake ID. Empty when the parent game could not be resolved." example:"1091918939407908944"`
	Position            int        `json:"position" description:"Prize rank within the game."`
	PrizeLabel          string     `json:"prizeLabel"`
	WinnerName          string     `json:"winnerName" description:"Inconnu when the player record could not be resolved."`
	WinningTicketNumber int        `json:"winningTicketNumber"`
	DrawnAt             *time.Time `json:"drawnAt" description:"Null when the prize has no recorded draw date."`
}

type LotoWinnersResponse struct {
	Winners []LotoWinner `json:"winners" nullable:"false"`
}

type APIError struct {
	Error string `json:"error" description:"Human-readable error message." example:"internal server error"`
}

func (LotoController) GetStats(c fiber.Ctx) error {
	db, err := lotoEntClient()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIError{Error: "database unavailable"})
	}

	ctx := c.Context()
	query := db.LotoGame.Query()

	if gameID := c.Query("game"); gameID != "" {
		query = query.Where(lotogame.ID(gameID))
	} else if c.Query("all") != "true" {
		query = query.Where(lotogame.Active(true))
	}
	if guildID := c.Query("guild"); guildID != "" {
		query = query.Where(lotogame.GuildID(guildID))
	}

	games, err := query.Order(ent.Desc(lotogame.FieldCreatedAt)).All(ctx)
	if err != nil {
		logger.Error("api: loading loto games", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: "internal server error"})
	}

	response := LotoStatsResponse{Games: make([]LotoGameStats, 0, len(games))}
	if len(games) == 0 {
		return c.JSON(response)
	}

	ids := make([]string, 0, len(games))
	for _, g := range games {
		ids = append(ids, g.ID)
	}

	ticketsPerGame, err := lotoCountTicketsPerGame(ctx, db, ids)
	if err != nil {
		logger.Error("api: counting loto tickets", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: "internal server error"})
	}
	playersPerGame, err := lotoCountPlayersPerGame(ctx, db, ids)
	if err != nil {
		logger.Error("api: counting loto players", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: "internal server error"})
	}
	prizesPerGame, err := lotoCountPrizesPerGame(ctx, db, ids)
	if err != nil {
		logger.Error("api: counting loto prizes", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: "internal server error"})
	}

	for _, g := range games {
		sold := ticketsPerGame[g.ID]
		stats := LotoGameStats{
			ID:           g.ID,
			GuildID:      g.GuildID,
			Name:         g.Name,
			Active:       g.Active,
			TicketPrice:  g.TicketPrice,
			TicketsSold:  sold,
			Pot:          sold * g.TicketPrice,
			PlayersCount: playersPerGame[g.ID],
			PrizesCount:  prizesPerGame[g.ID],
			CreatedAt:    g.CreatedAt,
		}
		response.Games = append(response.Games, stats)
		response.Totals.TicketsSold += stats.TicketsSold
		response.Totals.Pot += stats.Pot
		response.Totals.Players += stats.PlayersCount
	}
	response.Totals.Games = len(response.Games)

	return c.JSON(response)
}

func (LotoController) GetWinners(c fiber.Ctx) error {
	db, err := lotoEntClient()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIError{Error: "database unavailable"})
	}

	limit := lotoMaxWinnersLimit
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(APIError{Error: "invalid limit"})
		}
		limit = min(n, lotoMaxWinnersLimit)
	}

	ctx := c.Context()
	query := db.LotoPrize.Query().Where(lotoprize.HasWinner())

	if gameID := c.Query("game"); gameID != "" {
		query = query.Where(lotoprize.GameID(gameID))
	}
	if guildID := c.Query("guild"); guildID != "" {
		query = query.Where(lotoprize.HasGameWith(lotogame.GuildID(guildID)))
	}

	prizes, err := query.
		Order(ent.Desc(lotoprize.FieldDrawnAt), ent.Asc(lotoprize.FieldPosition)).
		Limit(limit).
		All(ctx)
	if err != nil {
		logger.Error("api: loading loto winners", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: "internal server error"})
	}

	response := LotoWinnersResponse{Winners: make([]LotoWinner, 0, len(prizes))}
	if len(prizes) == 0 {
		return c.JSON(response)
	}

	gameIDs := make([]string, 0, len(prizes))
	playerIDs := make([]string, 0, len(prizes))
	seenGames := make(map[string]struct{}, len(prizes))
	seenPlayers := make(map[string]struct{}, len(prizes))
	for _, p := range prizes {
		if _, ok := seenGames[p.GameID]; !ok {
			seenGames[p.GameID] = struct{}{}
			gameIDs = append(gameIDs, p.GameID)
		}
		if p.WinnerPlayerID == "" {
			continue
		}
		if _, ok := seenPlayers[p.WinnerPlayerID]; !ok {
			seenPlayers[p.WinnerPlayerID] = struct{}{}
			playerIDs = append(playerIDs, p.WinnerPlayerID)
		}
	}

	games, err := db.LotoGame.Query().Where(lotogame.IDIn(gameIDs...)).All(ctx)
	if err != nil {
		logger.Error("api: loading loto games for winners", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: "internal server error"})
	}
	lotoGamesByID := make(map[string]*ent.LotoGame, len(games))
	for _, g := range games {
		lotoGamesByID[g.ID] = g
	}

	playerNames := make(map[string]string, len(playerIDs))
	if len(playerIDs) > 0 {
		players, err := db.LotoPlayer.Query().Where(lotoplayer.IDIn(playerIDs...)).All(ctx)
		if err != nil {
			logger.Error("api: loading loto players for winners", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: "internal server error"})
		}
		for _, p := range players {
			playerNames[p.ID] = p.Name
		}
	}

	for _, p := range prizes {
		winner := LotoWinner{
			GameID:              p.GameID,
			Position:            p.Position,
			PrizeLabel:          p.Label,
			WinnerName:          "Inconnu",
			WinningTicketNumber: p.WinningTicketNumber,
		}
		if game, ok := lotoGamesByID[p.GameID]; ok {
			winner.GameName = game.Name
			winner.GuildID = game.GuildID
		}
		if name, ok := playerNames[p.WinnerPlayerID]; ok && name != "" {
			winner.WinnerName = name
		}
		if !p.DrawnAt.IsZero() {
			drawnAt := p.DrawnAt
			winner.DrawnAt = &drawnAt
		}
		response.Winners = append(response.Winners, winner)
	}

	return c.JSON(response)
}

func lotoEntClient() (*ent.Client, error) {
	if database.Default == nil {
		return nil, errLotoNoDatabase
	}
	return database.Default.Ent(), nil
}

type lotoNoDatabaseError struct{}

func (lotoNoDatabaseError) Error() string { return "database not initialised" }

var errLotoNoDatabase = lotoNoDatabaseError{}

func lotoCountTicketsPerGame(ctx context.Context, db *ent.Client, gameIDs []string) (map[string]int, error) {
	tickets, err := db.LotoTicket.Query().
		Where(lototicket.GameIDIn(gameIDs...)).
		Select(lototicket.FieldGameID).
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(gameIDs))
	for _, id := range tickets {
		counts[id]++
	}
	return counts, nil
}

func lotoCountPlayersPerGame(ctx context.Context, db *ent.Client, gameIDs []string) (map[string]int, error) {
	players, err := db.LotoPlayer.Query().
		Where(lotoplayer.GameIDIn(gameIDs...)).
		Select(lotoplayer.FieldGameID).
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(gameIDs))
	for _, id := range players {
		counts[id]++
	}
	return counts, nil
}

func lotoCountPrizesPerGame(ctx context.Context, db *ent.Client, gameIDs []string) (map[string]int, error) {
	prizes, err := db.LotoPrize.Query().
		Where(lotoprize.GameIDIn(gameIDs...)).
		Select(lotoprize.FieldGameID).
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int, len(gameIDs))
	for _, id := range prizes {
		counts[id]++
	}
	return counts, nil
}
