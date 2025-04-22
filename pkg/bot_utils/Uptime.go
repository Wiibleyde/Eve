package bot_utils

import "time"

var StartTime time.Time

func InitStartTime() {
	StartTime = time.Now()
}
