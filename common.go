package main

import (
	"strings"
	"time"
)

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Nicks(nicksRaw string) map[string]bool {
	var nicksM map[string]bool
	if len(nicksRaw) > 0 {
		nicksS := strings.Split(nicksRaw, ",")
		nicksM = make(map[string]bool, 0)
		for _, nick := range nicksS {
			nicksM[strings.Trim(nick, " ")] = true
			//dr.nicks[projectHrefPrefix+strings.ToLower(k)] = v
		}
	}

	return nicksM
}

func StartDate(startDateRaw string) time.Time {
	if startDateRaw == "" {
		return time.Unix(0, 0)
	}
	startTime, err := time.Parse("02.01.2006", startDateRaw)
	if err != nil {
		handleError(err, "Ошибка разбора даты начала")
	}
	return startTime
}
