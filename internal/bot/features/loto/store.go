package loto

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/lotogame"
	"Eve/internal/database/ent/lotoplayer"
	"Eve/internal/database/ent/lotoprize"
	"Eve/internal/database/ent/lototicket"

	"github.com/google/uuid"
)

var (
	errPlayerNotFound  = errors.New("loto: player not found")
	errNameTaken       = errors.New("loto: player name already used")
	errNoTickets       = errors.New("loto: no ticket to remove")
	errDrawAlreadyDone = errors.New("loto: draw already performed")
)

type snapshot struct {
	game    *ent.LotoGame
	players []*ent.LotoPlayer
	tickets []*ent.LotoTicket
	prizes  []*ent.LotoPrize
}

func client() *ent.Client {
	return database.Default.Ent()
}

func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return fmt.Errorf("%w (rollback: %v)", err, rerr)
	}
	return err
}

func activeGame(ctx context.Context, guildID string) (*ent.LotoGame, error) {
	game, err := client().LotoGame.Query().
		Where(lotogame.GuildID(guildID), lotogame.Active(true)).
		Order(ent.Desc(lotogame.FieldCreatedAt)).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return game, nil
}

func gameByID(ctx context.Context, id string) (*ent.LotoGame, error) {
	game, err := client().LotoGame.Get(ctx, id)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return game, nil
}

func loadSnapshot(ctx context.Context, gameID string) (*snapshot, error) {
	game, err := gameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, nil
	}

	players, err := client().LotoPlayer.Query().
		Where(lotoplayer.GameID(gameID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	tickets, err := client().LotoTicket.Query().
		Where(lototicket.GameID(gameID)).
		Order(ent.Asc(lototicket.FieldNumber)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	prizes, err := client().LotoPrize.Query().
		Where(lotoprize.GameID(gameID)).
		Order(ent.Asc(lotoprize.FieldPosition)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return &snapshot{game: game, players: players, tickets: tickets, prizes: prizes}, nil
}

func createGame(ctx context.Context, guildID, name string, ticketPrice, cooldownMinutes, maxTickets int, prizeLabels []string) (*snapshot, error) {
	tx, err := client().Tx(ctx)
	if err != nil {
		return nil, err
	}

	create := tx.LotoGame.Create().
		SetID(uuid.NewString()).
		SetGuildID(guildID).
		SetName(name).
		SetActive(true).
		SetTicketPrice(ticketPrice).
		SetCooldownMinutes(cooldownMinutes)
	if maxTickets > 0 {
		create.SetMaxTicketsPerPurchase(maxTickets)
	}

	game, err := create.Save(ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}

	builders := make([]*ent.LotoPrizeCreate, 0, len(prizeLabels))
	for i, label := range prizeLabels {
		builders = append(builders, tx.LotoPrize.Create().
			SetID(uuid.NewString()).
			SetGameID(game.ID).
			SetLabel(label).
			SetPosition(i))
	}

	prizes, err := tx.LotoPrize.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, rollback(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &snapshot{game: game, prizes: prizes}, nil
}

func setGameMessage(ctx context.Context, gameID, channelID, messageID string) error {
	return client().LotoGame.UpdateOneID(gameID).
		SetChannelID(channelID).
		SetMessageID(messageID).
		Exec(ctx)
}

func findPlayer(ctx context.Context, gameID, name string) (*ent.LotoPlayer, error) {
	player, err := client().LotoPlayer.Query().
		Where(lotoplayer.GameID(gameID), lotoplayer.Name(name)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return player, nil
}

func addTickets(ctx context.Context, gameID string, player *ent.LotoPlayer, name, sellerID string, count int) error {
	now := time.Now()

	tx, err := client().Tx(ctx)
	if err != nil {
		return err
	}

	if player == nil {
		player, err = tx.LotoPlayer.Create().
			SetID(uuid.NewString()).
			SetGameID(gameID).
			SetName(name).
			SetLastPlay(now).
			Save(ctx)
	} else {
		player, err = tx.LotoPlayer.UpdateOneID(player.ID).
			SetLastPlay(now).
			Save(ctx)
	}
	if err != nil {
		return rollback(tx, err)
	}

	var next int
	last, err := tx.LotoTicket.Query().
		Where(lototicket.GameID(gameID)).
		Order(ent.Desc(lototicket.FieldNumber)).
		First(ctx)
	switch {
	case err == nil:
		next = last.Number + 1
	case ent.IsNotFound(err):
		next = 1
	default:
		return rollback(tx, err)
	}

	builders := make([]*ent.LotoTicketCreate, 0, count)
	for i := 0; i < count; i++ {
		builders = append(builders, tx.LotoTicket.Create().
			SetID(uuid.NewString()).
			SetGameID(gameID).
			SetPlayerID(player.ID).
			SetSellerID(sellerID).
			SetNumber(next+i))
	}

	if _, err := tx.LotoTicket.CreateBulk(builders...).Save(ctx); err != nil {
		return rollback(tx, err)
	}

	return tx.Commit()
}

func removeTickets(ctx context.Context, gameID, playerName string, count int, sellerFilter string) (int, error) {
	player, err := findPlayer(ctx, gameID, playerName)
	if err != nil {
		return 0, err
	}
	if player == nil {
		return 0, errPlayerNotFound
	}

	query := client().LotoTicket.Query().
		Where(lototicket.GameID(gameID), lototicket.PlayerID(player.ID))
	if sellerFilter != "" {
		query = query.Where(lototicket.SellerID(sellerFilter))
	}

	tickets, err := query.
		Order(ent.Desc(lototicket.FieldNumber)).
		Limit(count).
		All(ctx)
	if err != nil {
		return 0, err
	}
	if len(tickets) == 0 {
		return 0, errNoTickets
	}

	ids := make([]string, 0, len(tickets))
	for _, t := range tickets {
		ids = append(ids, t.ID)
	}

	return client().LotoTicket.Delete().Where(lototicket.IDIn(ids...)).Exec(ctx)
}

func renamePlayer(ctx context.Context, gameID, oldName, newName string) error {
	player, err := findPlayer(ctx, gameID, oldName)
	if err != nil {
		return err
	}
	if player == nil {
		return errPlayerNotFound
	}

	taken, err := findPlayer(ctx, gameID, newName)
	if err != nil {
		return err
	}
	if taken != nil {
		return errNameTaken
	}

	err = client().LotoPlayer.UpdateOneID(player.ID).SetName(newName).Exec(ctx)
	if ent.IsConstraintError(err) {
		return errNameTaken
	}
	return err
}

type assignment struct {
	prize      *ent.LotoPrize
	playerID   string
	playerName string
	ticket     *ent.LotoTicket
}

func applyDraw(ctx context.Context, gameID string, assignments []assignment) error {
	now := time.Now()

	tx, err := client().Tx(ctx)
	if err != nil {
		return err
	}

	closed, err := tx.LotoGame.Update().
		Where(lotogame.ID(gameID), lotogame.Active(true)).
		SetActive(false).
		Save(ctx)
	if err != nil {
		return rollback(tx, err)
	}
	if closed == 0 {
		return rollback(tx, errDrawAlreadyDone)
	}

	for _, a := range assignments {
		err := tx.LotoPrize.UpdateOneID(a.prize.ID).
			SetWinnerPlayerID(a.playerID).
			SetWinningTicketNumber(a.ticket.Number).
			SetDrawnAt(now).
			Exec(ctx)
		if err != nil {
			return rollback(tx, err)
		}
	}

	return tx.Commit()
}

func deleteGame(ctx context.Context, gameID string) error {
	return client().LotoGame.DeleteOneID(gameID).Exec(ctx)
}
