package utils

import "time"

func GetStartDateForPeriod(period string) time.Time {
	now := time.Now().UTC()
    switch period {
    case "week":
        return now.AddDate(0, 0, -7)
    case "month":
        return now.AddDate(0, -1, 0)
    case "30days":
        return now.AddDate(0, 0, -30)
    case "today":
        year, month, day := now.Date()
        return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
    default:
        return now.AddDate(0, 0, -7)
    }
}

func GetStartDateFromDays(days string) time.Time {
	now := time.Now().UTC()

	switch days {
	case "7days":
		return now.AddDate(0, 0, -7)
	case "30days":
		return now.AddDate(0, 0, -30)
	case "90days":
		return now.AddDate(0, 0, -90)
	default:
		return now.AddDate(0, 0, -7)
	}
}
