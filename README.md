![travis build](https://travis-ci.org/RoboCup-SSL/ssl-go-tools.svg?branch=master "travis build status")
[![Go Report Card](https://goreportcard.com/badge/github.com/RoboCup-SSL/ssl-go-tools?style=flat-square)](https://goreportcard.com/report/github.com/RoboCup-SSL/ssl-go-tools)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/RoboCup-SSL/ssl-go-tools)

# ssl-go-tools

Collection of packages to do common stuff for the RoboCup SSL league like reading, writing, sending, receiving and parsing messages.

## Requirements
You need to install following dependencies first: 
 * Go >= 1.9
 
## Installation

Use go get to install all packages / executables:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/...
```

## Run
The executables are installed to your $GOPATH/bin folder. If you have it on your $PATH, you can directly run them. Else, switch to this folder first.

## Usage and further details

Have a look at the individual packages and their containing READMEs:

 * [ssl-recorder](cmd/ssl-recorder/README.md)
 * [ssl-player](cmd/ssl-player/README.md)
 * [ssl-logcutter](cmd/ssl-logcutter/README.md)
 * [ssl-logstats](cmd/ssl-logstats/README.md)
 * [matchduration](cmd/matchduration/README.md)
 * [numcards](cmd/numcards/README.md)
