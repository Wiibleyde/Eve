package birthday

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/maintenance"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"
	"Eve/internal/logger"
)

func StartScheduler(bot *bot.Client) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))

			if maintenance.Enabled() {
				logger.Debug("Skipping birthday tick: maintenance mode enabled")
				continue
			}

			birthdays, err := getTodayBirthdays(context.Background())
			if err != nil {
				logger.Error("fetching today birthdays: " + err.Error())
				continue
			}
			for _, b := range birthdays {
				sendBirthdayMessage(bot, b)
			}
		}
	}()
}

func getTodayBirthdays(ctx context.Context) ([]*ent.Birthday, error) {
	today := time.Now()
	return database.Default.Ent().Birthday.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"EXTRACT(MONTH FROM birthday) = $1 AND EXTRACT(DAY FROM birthday) = $2",
				int(today.Month()), today.Day(),
			))
		}).
		All(ctx)
}

func sendBirthdayMessage(client *bot.Client, b *ent.Birthday) {
	ctx := context.Background()

	configs, err := database.Default.Ent().GuildConfig.Query().
		Where(guildconfig.Key(tables.BirthdayChannel.String())).
		All(ctx)
	if err != nil {
		logger.Error("fetching birthday channel configs: " + err.Error())
		return
	}

	userID, err := snowflake.Parse(b.DiscordID)
	if err != nil {
		logger.Error("parsing user ID: " + err.Error())
		return
	}

	for _, cfg := range configs {
		guildID, err := snowflake.Parse(cfg.GuildID)
		if err != nil {
			logger.Error("parsing guild ID: " + err.Error())
			continue
		}

		if _, err := client.Rest.GetMember(guildID, userID); err != nil {
			continue
		}

		channelID, err := snowflake.Parse(cfg.Value)
		if err != nil {
			logger.Error("parsing channel ID: " + err.Error())
			continue
		}

		embed := embeds.BaseEmbed()
		embed.Title = "Joyeux anniversaire !"
		embed.Description = fmt.Sprintf("Aujourd'hui c'est l'anniversaire de <@%s> ! 🥳", b.DiscordID)

		if _, err := client.Rest.CreateMessage(channelID, discord.MessageCreate{
			Embeds: []discord.Embed{embed},
		}); err != nil {
			logger.Error("sending birthday message: " + err.Error())
		}
	}
}
