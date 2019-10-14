package matchstats

import (
	"encoding/csv"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/pkg/errors"
	"os"
	"sort"
	"strconv"
	"strings"
)

func writeCsv(header []string, data [][]string, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create CSV output file")
	}

	if _, err := f.WriteString("#" + strings.Join(header, ",") + "\n"); err != nil {
		return err
	}

	w := csv.NewWriter(f)
	if err := w.WriteAll(data); err != nil {
		return err
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}

	return f.Close()
}

func WriteGamePhaseDurations(matchStatsCollection *sslproto.MatchStatsCollection, filename string) error {

	header := []string{"File"}
	for i := 0; i < len(sslproto.GamePhaseType_name); i++ {
		header = append(header, sslproto.GamePhaseType_name[int32(i)])
	}

	var records [][]string
	for _, matchStats := range matchStatsCollection.MatchStats {
		record := []string{matchStats.Name}
		for i := 0; i < len(sslproto.GamePhaseType_name); i++ {
			name := sslproto.GamePhaseType_name[int32(i)]
			record = append(record, strconv.FormatUint(uint64(matchStats.TimePerGamePhase[name]), 10))
		}
		records = append(records, record)
	}

	return writeCsv(header, records, filename)
}

func WriteTeamMetricsPerGame(matchStatsCollection *sslproto.MatchStatsCollection, filename string) error {

	header := []string{"File", "Extra time", "Shootout", "Team", "Goals", "Fouls", "Yellow Cards", "Red Cards", "Timeout Time", "Penalty Shots"}

	var records [][]string
	for _, matchStats := range matchStatsCollection.MatchStats {
		recordYellow := []string{matchStats.Name, strconv.FormatBool(matchStats.ExtraTime), strconv.FormatBool(matchStats.Shootout)}
		recordYellow = append(recordYellow, teamNumbers(matchStats.TeamStatsYellow)...)
		records = append(records, recordYellow)
		recordBlue := []string{matchStats.Name, strconv.FormatBool(matchStats.ExtraTime), strconv.FormatBool(matchStats.Shootout)}
		recordBlue = append(recordBlue, teamNumbers(matchStats.TeamStatsBlue)...)
		records = append(records, recordBlue)
	}

	return writeCsv(header, records, filename)
}

func WriteTeamMetricsSum(matchStatsCollection *sslproto.MatchStatsCollection, filename string) error {

	header := []string{"Team", "Goals", "Fouls", "Yellow Cards", "Red Cards", "Timeout Time", "Penalty Shots"}

	teams := map[string]*sslproto.TeamStats{}
	for _, matchStats := range matchStatsCollection.MatchStats {
		teams[matchStats.TeamStatsYellow.Name] = &sslproto.TeamStats{Name: matchStats.TeamStatsYellow.Name}
		teams[matchStats.TeamStatsBlue.Name] = &sslproto.TeamStats{Name: matchStats.TeamStatsBlue.Name}
	}

	for _, matchStats := range matchStatsCollection.MatchStats {
		addTeamStats(matchStats.TeamStatsYellow, teams[matchStats.TeamStatsYellow.Name])
		addTeamStats(matchStats.TeamStatsBlue, teams[matchStats.TeamStatsBlue.Name])
	}

	var teamNamesSorted []string
	for k := range teams {
		teamNamesSorted = append(teamNamesSorted, k)
	}
	sort.Strings(teamNamesSorted)

	var records [][]string
	for _, teamName := range teamNamesSorted {
		teamStats := teams[teamName]
		records = append(records, teamNumbers(teamStats))
	}

	return writeCsv(header, records, filename)
}

func addTeamStats(from *sslproto.TeamStats, to *sslproto.TeamStats) {
	to.Goals += from.Goals
	to.Fouls += from.Fouls
	to.YellowCards += from.YellowCards
	to.RedCards += from.RedCards
	to.TimeoutTime += from.TimeoutTime
	to.PenaltyShotsTotal += from.PenaltyShotsTotal
}

func teamNumbers(stats *sslproto.TeamStats) []string {
	return []string{
		stats.Name,
		uintToStr(stats.Goals),
		uintToStr(stats.Fouls),
		uintToStr(stats.YellowCards),
		uintToStr(stats.RedCards),
		uintToStr(stats.TimeoutTime),
		uintToStr(stats.PenaltyShotsTotal),
	}
}

func uintToStr(n uint32) string {
	return strconv.FormatUint(uint64(n), 10)
}
