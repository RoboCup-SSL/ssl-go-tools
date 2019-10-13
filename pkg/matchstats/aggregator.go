package matchstats

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"os"
)

type Aggregator struct {
	Collection sslproto.MatchStatsCollection
}

func NewAggregator() *Aggregator {
	generator := new(Aggregator)
	generator.Collection = sslproto.MatchStatsCollection{}
	generator.Collection.MatchStats = []*sslproto.MatchStats{}
	return generator
}

func (a *Aggregator) Process(filename string) error {
	generator := NewGenerator()

	matchStats, err := generator.Process(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create match states")
	} else {
		a.Collection.MatchStats = append(a.Collection.MatchStats, matchStats)
	}
	return nil
}

func (a *Aggregator) WriteJson(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create JSON output file")
	}

	jsonMarsh := jsonpb.Marshaler{EmitDefaults: true, Indent: "  "}
	err = jsonMarsh.Marshal(f, &a.Collection)
	if err != nil {
		return errors.Wrap(err, "Could not marshal match stats to json")
	}
	return f.Close()
}

func (a *Aggregator) WriteBin(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create Binary output file")
	}

	bytes, err := proto.Marshal(&a.Collection)
	if err != nil {
		return errors.Wrap(err, "Could not marshal match stats to binary")
	}
	_, err = f.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "Could not write match stats to binary")
	}
	return f.Close()
}
