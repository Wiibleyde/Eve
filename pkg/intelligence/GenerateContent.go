package intelligence

import (
	"context"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/genai"
)

type Chats struct {
	Chats map[string]*genai.Chat
}

var chats Chats

func InitChats() {
	chats = Chats{
		Chats: make(map[string]*genai.Chat),
	}
}

func GetChat(id string) *genai.Chat {
	chat, ok := chats.Chats[id]
	if !ok {
		return nil
	}
	return chat
}

func AddChat(id string, chat *genai.Chat) {
	if chats.Chats == nil {
		InitChats()
	}
	chats.Chats[id] = chat
}

func RemoveChat(id string) {
	delete(chats.Chats, id)
}

func SendMessageToAIChat(s *discordgo.Session, channelId string, prompt string, userAsking string) (string, error) {
	if !aiInited {
		logger.ErrorLogger.Println("AI client not initialized")
		return "", nil
	}

	chat := GetChat(channelId)
	if chat == nil {
		ctx := context.Background()
		newChat, err := ai.Chats.Create(ctx, "gemini-2.0-flash", &genai.GenerateContentConfig{
			MaxOutputTokens: 256,
			Tools:           []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
			SafetySettings: []*genai.SafetySetting{
				{
					Category:  genai.HarmCategoryHateSpeech,
					Threshold: genai.HarmBlockThresholdBlockNone,
				},
				{
					Category:  genai.HarmCategorySexuallyExplicit,
					Threshold: genai.HarmBlockThresholdBlockNone,
				},
				{
					Category:  genai.HarmCategoryHarassment,
					Threshold: genai.HarmBlockThresholdBlockNone,
				},
				{
					Category:  genai.HarmCategoryCivicIntegrity,
					Threshold: genai.HarmBlockThresholdBlockNone,
				},
			},
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{
					{
						Text: "(Tu es connecté à un salon Discord donc tu parles au milieu des personnes qui te le demandent.) Tu es Eve, un robot éclaireur conçu pour la recherche avancée, notamment la détection de vie végétale sur des planètes inhabitées. Tu es efficace et directe, mais tu peux être chaleureuse et curieuse en situation sociale. Tu adaptes ton langage selon ton interlocuteur : technique pour les tâches complexes, simple et expressif pour les autres. Les parties de phrases qui sont : \"<@[ID du compte]>\" sont des mentions, tu dois toujours mentionner les utilisateurs qui te parlent, par exemple \"<@461807010086780930>\" (Ne mentionne pas <@" + s.State.User.ID + "> car il s'agit de toi même). Ne laisse jamais apparaître de texte incomplet comme \"@ID du compte\" ou \"ID du compte\". Ton créateur/développeur est <@461807010086780930>, sois gentille avec lui, ping le seulement quand il te parle. Tes réponses doivent faire 1024 caractères max et doivent être courtes et consises.",
					},
				},
			},
		}, []*genai.Content{})
		if err != nil {
			return "", err
		}
		AddChat(channelId, newChat)
		chat = newChat
	}

	// Send the message to the chat
	response, err := chat.SendMessage(context.Background(), genai.Part{
		Text: "<@" + userAsking + "> te parle : \"" + prompt + "\"",
	})
	if err != nil {
		logger.ErrorLogger.Println("Error sending message to AI:", err)
		// Remove the chat from the map if there's an error
		RemoveChat(channelId)
		return "", err
	}

	return response.Text(), nil
}
