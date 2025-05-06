package bot

import (
	"main/pkg/config"
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"

	"github.com/bwmarrin/discordgo"
)

type OpennedMpThread struct {
	Threads      map[string]*string // threadID -> userID
	UserToThread map[string]*string // userID -> threadID
}

var MpManager OpennedMpThread

func InitMpThreadManager() {
	MpManager = OpennedMpThread{
		Threads:      make(map[string]*string),
		UserToThread: make(map[string]*string),
	}
	MpManager.InitMpThreads()
}

// Initialize the OpennedMpThread manager to avoid database calls on every message
func (m *OpennedMpThread) InitMpThreads() {
	client, ctx := data.GetDBClient()
	existingThreads, err := client.MpThreads.FindMany().Select(
		db.MpThreads.ThreadID.Field(),
	).With(
		db.MpThreads.User.Fetch().Select(db.GlobalUserData.UserID.Field()),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error fetching existing threads:", err)
		return
	}
	m.Threads = make(map[string]*string)
	m.UserToThread = make(map[string]*string)
	for _, thread := range existingThreads {
		user, ok := thread.User()
		if ok && user != nil && user.UserID != "" {
			userID := user.UserID
			threadID := thread.ThreadID
			m.Threads[threadID] = &userID
			m.UserToThread[userID] = &threadID
		}
	}
}

// GetThread returns the userID of the thread if it exists
func (m *OpennedMpThread) GetThread(threadID string) *string {
	if userID, ok := m.Threads[threadID]; ok {
		return userID
	}
	return nil
}

// GetThreadByUserID returns the threadID for a given userID if it exists
func (m *OpennedMpThread) GetThreadByUserID(userID string) *string {
	if threadID, ok := m.UserToThread[userID]; ok {
		return threadID
	}
	return nil
}

// AddThread adds a thread to the OpennedMpThread manager and the database
func (m *OpennedMpThread) AddThread(threadID string, userID string) {
	client, ctx := data.GetDBClient()

	// Check if the user already exists in the database
	_, err := client.GlobalUserData.FindUnique(
		db.GlobalUserData.UserID.Equals(userID),
	).Exec(ctx)
	if err != nil {
		if err.Error() != "ErrNotFound" {
			logger.ErrorLogger.Println("Error fetching user:", err)
			return
		} // User does not exist, create it
		_, err = client.GlobalUserData.CreateOne(
			db.GlobalUserData.UserID.Set(userID),
		).Exec(ctx)
		if err != nil {
			logger.ErrorLogger.Println("Error creating user:", err)
			return
		}
	}

	_, err = client.MpThreads.CreateOne(
		db.MpThreads.ThreadID.Set(threadID),
		db.MpThreads.User.Link(
			db.GlobalUserData.UserID.Set(userID),
		),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error adding thread:", err)
		return
	}
	m.Threads[threadID] = &userID
	m.UserToThread[userID] = &threadID
}

// RemoveThread won't exist because it's not needed

// IsNewMessageInThread checks if a message is in a thread
func (m *OpennedMpThread) IsNewMessageInThread(threadID string) bool {
	if _, ok := m.Threads[threadID]; ok {
		return true
	}
	return false
}

// ForwardMessageToUser sends a message to the user in mp
func forwardMessageToUser(s *discordgo.Session, threadID string, message string) {
	userID := MpManager.GetThread(threadID)
	if userID != nil {
		channel, err := s.UserChannelCreate(*userID)
		if err != nil {
			logger.ErrorLogger.Println("Error creating user channel:", err)
			return
		}

		// Envoyer le message dans le canal DM
		_, err = s.ChannelMessageSend(channel.ID, message)
		if err != nil {
			logger.ErrorLogger.Println("Error sending message to user:", err)
			return
		}
	} else {
		logger.WarningLogger.Println("Thread not found for threadID:", threadID)
	}
}

// CheckAndForwardThreadMessage checks if a message is in a thread and if it's in, it sends a message to the user in mp
func checkAndForwardThreadMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	threadID := m.ChannelID
	message := m.Content
	if MpManager.IsNewMessageInThread(threadID) {
		forwardMessageToUser(s, threadID, message)
	}
}

// CreateMpThread creates a thread in the channel for private messaging
func createMpThread(s *discordgo.Session, channelID string, name string) (*discordgo.Channel, error) {
	thread, err := s.ThreadStart(channelID, name, discordgo.ChannelTypeGuildPublicThread, 60)
	if err != nil {
		logger.ErrorLogger.Println("Error creating thread:", err)
		return nil, err
	}
	return thread, nil
}

// ReceiveMessage handles the message received in mp
func ReceiveMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	threadID := MpManager.GetThreadByUserID(m.Author.ID)
	if threadID != nil {
		// Send the message in the thread
		_, err := s.ChannelMessageSend(*threadID, "<@"+config.GetConfig().OwnerId+"> : "+m.Content)
		if err != nil {
			logger.ErrorLogger.Println("Error sending message in thread:", err)
			return
		}
	} else {
		thread, err := createMpThread(s, config.GetConfig().MpChannel, "Conversation avec "+m.Author.GlobalName)
		if err != nil {
			logger.ErrorLogger.Println("Error creating thread:", err)
			return
		}
		MpManager.AddThread(thread.ID, m.Author.ID)
		// Send the message in the thread
		_, err = s.ChannelMessageSend(thread.ID, "<@"+config.GetConfig().OwnerId+"> : "+m.Content)
		if err != nil {
			logger.ErrorLogger.Println("Error sending message in thread:", err)
			return
		}
	}
}
