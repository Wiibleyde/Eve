package data

import (
	"main/pkg/logger"
	"main/prisma/db"
)

// Logs_Insert adds a new log entry to the database with the specified level and message
func Logs_Insert(level string, logMessage string) (*db.LogsModel, error) {
	client, ctx := GetDBClient()

	// First get the level ID
	logLevel, err := client.LogLevel.FindFirst(
		db.LogLevel.Name.Equals(level),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error finding log level:", err)
		return nil, err
	}

	// Create the log entry
	log, err := client.Logs.CreateOne(
		db.Logs.Message.Set(logMessage),
		db.Logs.Level.Link(
			db.LogLevel.ID.Equals(logLevel.ID),
		),
	).Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error creating log:", err)
		return nil, err
	}

	return log, nil
}

// Logs_GetById retrieves a log entry by its ID
func Logs_GetById(id int) (*db.LogsModel, error) {
	client, ctx := GetDBClient()

	log, err := client.Logs.FindUnique(
		db.Logs.ID.Equals(id),
	).With(
		db.Logs.Level.Fetch(),
	).Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error finding log:", err)
		return nil, err
	}

	return log, nil
}

// Logs_GetAll retrieves all log entries from the database
func Logs_GetAll() ([]db.LogsModel, error) {
	client, ctx := GetDBClient()

	logs, err := client.Logs.FindMany().With(
		db.Logs.Level.Fetch(),
	).Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error finding logs:", err)
		return nil, err
	}

	return logs, nil
}

// Logs_GetByLevel retrieves all log entries with a specific level
func Logs_GetByLevel(level string) ([]db.LogsModel, error) {
	client, ctx := GetDBClient()

	// First get the level ID
	logLevel, err := client.LogLevel.FindFirst(
		db.LogLevel.Name.Equals(level),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error finding log level:", err)
		return nil, err
	}

	logs, err := client.Logs.FindMany(
		db.Logs.LevelID.Equals(logLevel.ID),
	).With(
		db.Logs.Level.Fetch(),
	).Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error finding logs:", err)
		return nil, err
	}

	return logs, nil
}

// Logs_Update updates a log entry's message
func Logs_Update(id int, newMessage string) (*db.LogsModel, error) {
	client, ctx := GetDBClient()

	log, err := client.Logs.FindUnique(
		db.Logs.ID.Equals(id),
	).Update(
		db.Logs.Message.Set(newMessage),
	).Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error updating log:", err)
		return nil, err
	}

	return log, nil
}

// Logs_Delete deletes a log entry by its ID
func Logs_Delete(id int) error {
	client, ctx := GetDBClient()

	_, err := client.Logs.FindUnique(
		db.Logs.ID.Equals(id),
	).Delete().Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error deleting log:", err)
		return err
	}

	return nil
}

// Logs_DeleteAll deletes all log entries
func Logs_DeleteAll() error {
	client, ctx := GetDBClient()

	_, err := client.Logs.FindMany().Delete().Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error deleting all logs:", err)
		return err
	}

	return nil
}

// Logs_DeleteByLevel deletes all log entries with a specific level
func Logs_DeleteByLevel(level string) error {
	client, ctx := GetDBClient()

	// First get the level ID
	logLevel, err := client.LogLevel.FindFirst(
		db.LogLevel.Name.Equals(level),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error finding log level:", err)
		return err
	}

	_, err = client.Logs.FindMany(
		db.Logs.LevelID.Equals(logLevel.ID),
	).Delete().Exec(ctx)

	if err != nil {
		logger.ErrorLogger.Println("Error deleting logs by level:", err)
		return err
	}

	return nil
}
