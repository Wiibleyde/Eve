package bot_utils

import "main/pkg/logger"

var isMaintenanceMode bool = false

func SetMaintenanceMode(mode bool) {
	isMaintenanceMode = mode
	logger.InfoLogger.Println("Maintenance mode set to:", mode)
}

func IsMaintenanceMode() bool {
	return isMaintenanceMode
}
