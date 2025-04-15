package data

import (
	"context"
	"main/pkg/logger"
	"main/prisma/db"
	"sync"
)

var (
	client     *db.PrismaClient
	ctx        context.Context
	clientOnce sync.Once
)

// GetDBClient initializes and returns a connected database client using a singleton pattern
func GetDBClient() (*db.PrismaClient, context.Context) {
	clientOnce.Do(func() {
		client = db.NewClient()
		err := client.Prisma.Connect()
		if err != nil {
			logger.ErrorLogger.Panicln("Error connecting to database,", err)
		}

		ctx = context.Background()
	})

	return client, ctx
}

// CloseDBConnection should be called when shutting down the application
func CloseDBConnection() {
	if client != nil {
		if err := client.Prisma.Disconnect(); err != nil {
			logger.ErrorLogger.Println("Error disconnecting from database:", err)
		}
	}
}

// InitDatabase is maintained for backward compatibility
func InitDatabase() {
	GetDBClient()
}
