# ssl-match-stats

Gather statistics from a set of [official SSL log files](https://ssl.robocup.org/game-logs/).

## Installation

Use go-get to install this executable:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/cmd/ssl-match-stats
```

## Usage

The binary is called `ssl-match-stats`.
Pass in a list of log files to be processed, e.g.: `ssl-match-stats *.log.gz`
