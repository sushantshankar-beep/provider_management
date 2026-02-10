package config

import (
	"os"
	"strconv"
)

func GetPayoutIntervalHours() int {
	hoursStr := os.Getenv("PAYOUT_INTERVAL_HOURS")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		return 6
	}
	return hours
}
