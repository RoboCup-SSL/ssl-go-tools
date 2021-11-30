[![CircleCI](https://circleci.com/gh/RoboCup-SSL/ssl-go-tools/tree/master.svg?style=svg)](https://circleci.com/gh/RoboCup-SSL/ssl-go-tools/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/RoboCup-SSL/ssl-go-tools?style=flat-square)](https://goreportcard.com/report/github.com/RoboCup-SSL/ssl-go-tools)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/RoboCup-SSL/ssl-go-tools)
[![Coverage](https://img.shields.io/badge/coverage-report-blue.svg)](https://circleci.com/api/v1.1/project/github/RoboCup-SSL/ssl-go-tools/latest/artifacts/0/coverage?branch=master)

# ssl-go-tools

Collection of packages to do common stuff for the RoboCup SSL league like reading, writing, sending, receiving and
parsing messages.

## Requirements

You need to install following dependencies first:

* Go >= 1.17

## Installation

Use go get to install all packages / executables:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/...
```

## Run

The executables are installed to your $GOPATH/bin folder. If you have it on your $PATH, you can directly run them. Else,
switch to this folder first.

## Usage and further details

Have a look at the individual packages and their containing READMEs:

* [ssl-auto-recorder](cmd/ssl-auto-recorder/README.md)
* [ssl-log-converter](cmd/ssl-log-converter/README.md)
* [ssl-log-cutter](cmd/ssl-log-cutter/README.md)
* [ssl-log-stats](cmd/ssl-log-stats/README.md)
* [ssl-log-player](cmd/ssl-log-player/README.md)
* [ssl-log-recorder](cmd/ssl-log-recorder/README.md)
* [ssl-vision-tracker-client](cmd/ssl-vision-tracker-client/README.md)
