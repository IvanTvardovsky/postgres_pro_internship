package utils

import (
	"fmt"
	"time"
)

func GetDate() string {
	now := time.Now()
	year, month, day := now.Date()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()
	date := fmt.Sprintf("%d-%d-%d %d:%d:%d", year, month, day, hour, minute, second)
	return date
}
