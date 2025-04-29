package commandHandler

import (
	"bytes"
	"encoding/json"
	"io"
	"main/pkg/bot_utils"
	"main/pkg/config"
	"main/pkg/logger"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const url = "https://www.blagues-api.fr/api/type/<TYPE>/random"

type Blague struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Joke   string `json:"joke"`
	Answer string `json:"answer"`
}

func BlagueHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	typeAsked := i.ApplicationCommandData().Options[0].StringValue()

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.ErrorLogger.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.GetConfig().BlagueApiToken)
	req.URL.Path = strings.Replace(req.URL.Path, "<TYPE>", typeAsked, 1)
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.ErrorLogger.Println("Error making request:", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.ErrorLogger.Println("Error response from API:", resp.Status)
		return err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorLogger.Println("Error reading response body:", err)
		return err
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var blague Blague
	err = json.NewDecoder(resp.Body).Decode(&blague)
	if err != nil {
		logger.ErrorLogger.Println("Error decoding response:", err)
		return err
	}

	embed := bot_utils.BasicEmbedBuilder(s)
	embed.Title = "Blague"
	embed.Description = "**" + blague.Joke + "**"
	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:  "Réponse",
			Value: "||" + blague.Answer + "||",
		},
	}
	embed.Footer.Text = "Eve et ses développeurs ne sont pas responsable des blagues affichées."
	embed.Color = 0x00FF00

	button := &discordgo.Button{
		Label:    "Afficher publiquement",
		Style:    discordgo.SuccessButton,
		CustomID: "jokeSetPublicButton",
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						button,
					},
				},
			},
		},
	})
	if err != nil {
		logger.ErrorLogger.Println("Error sending response:", err)
		return err
	}
	return nil
}
