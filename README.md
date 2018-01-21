# ssl-log-analyzer

Analyzing utility for log files from the RoboCup SSL.

# Install

Use go get to install all exe

```
# analyze match durations
go get github.com/g3force/ssl-log-analyzer/matchduration
# analyze number of cards given to teams
go get github.com/g3force/ssl-log-analyzer/numcards
```

# Run

```
matchduration <directory containing log files>
numcards <directory containing log files>
```
Log files may be compressed, if they end with `.gz`
