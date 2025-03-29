package auto

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"strings"
	"time"
)

func LogFileName(refMsg *gc.Referee, location *time.Location) string {
	teamNameYellow := strings.Replace(*refMsg.Yellow.Name, " ", "_", -1)
	teamNameBlue := strings.Replace(*refMsg.Blue.Name, " ", "_", -1)
	date := time.Unix(0, int64(*refMsg.PacketTimestamp*1000)).In(location).Format("2006-01-02_15-04")
	matchType := refMsg.GetMatchType().String()
	return fmt.Sprintf("%s_%s_%s-vs-%s.log", date, matchType, teamNameYellow, teamNameBlue)
}
