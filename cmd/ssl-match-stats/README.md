# ssl-match-stats

Gather statistics from a set of [official SSL log files](https://ssl.robocup.org/game-logs/).

## Installation

Use go-get to install this executable:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/cmd/ssl-match-stats
```

## Usage

The binary is called `ssl-match-stats`.

### Generate statistics from log files

Pass in a list of log files to be processed: `ssl-match-stats -generate *.log.gz`

### Export statistics to CSV files

First, generate the statistics with the command above. This will produce a `out.json` and `out.bin` file.

Run: `ssl-match-stats -exportCsv`

This will generate `*.csv` files that you can import in your favorite tool, like a spreadsheet tool.
