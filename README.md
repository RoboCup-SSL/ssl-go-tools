[![CircleCI](https://circleci.com/gh/RoboCup-SSL/ssl-go-tools/tree/master.svg?style=svg)](https://circleci.com/gh/RoboCup-SSL/ssl-go-tools/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/RoboCup-SSL/ssl-go-tools?style=flat-square)](https://goreportcard.com/report/github.com/RoboCup-SSL/ssl-go-tools)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/RoboCup-SSL/ssl-go-tools)
[![Coverage](https://img.shields.io/badge/coverage-report-blue.svg)](https://circleci.com/api/v1.1/project/github/RoboCup-SSL/ssl-go-tools/latest/artifacts/0/coverage?branch=master)

# ssl-go-tools

Collection of packages to do common stuff for the RoboCup SSL league like reading, writing, sending, receiving and
parsing messages.

## Installation from GitHub releases

The GitHub release page contains the latest stable binaries: https://github.com/RoboCup-SSL/ssl-go-tools/releases

Simply download the archive for your platform and extract to a folder of your choosing and run it from there.

## Installation with Go

If you have Go installed, you can install the tools with:

```shell
go install github.com/RoboCup-SSL/ssl-go-tools/...@latest
```

## Usage and further details

Have a look at the individual packages and their containing READMEs:

* [ssl-auto-recorder](cmd/ssl-auto-recorder/README.md)
* [ssl-log-converter](cmd/ssl-log-converter/README.md)
* [ssl-log-cutter](cmd/ssl-log-cutter/README.md)
* [ssl-log-stats](cmd/ssl-log-stats/README.md)
* [ssl-log-player](cmd/ssl-log-player/README.md)
* [ssl-log-recorder](cmd/ssl-log-recorder/README.md)
* [ssl-vision-tracker-client](cmd/ssl-vision-tracker-client/README.md)
