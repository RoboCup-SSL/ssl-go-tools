package matchstats

import (
	"encoding/csv"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
)

func WriteGamePhaseDurations(matchStatsCollection *sslproto.MatchStatsCollection, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create CSV output file")
	}

	header := []string{"Name"}
	for _, name := range sslproto.GamePhaseType_name {
		header = append(header, name)
	}

	if _, err := f.WriteString("#" + strings.Join(header, ",") + "\n"); err != nil {
		return err
	}

	w := csv.NewWriter(f)

	for _, matchStats := range matchStatsCollection.MatchStats {
		record := []string{matchStats.Name}
		for _, name := range sslproto.GamePhaseType_name {
			record = append(record, strconv.FormatUint(uint64(matchStats.TimePerGamePhase[name]), 10))
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}

	return f.Close()
}
