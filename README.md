# ssl-go-tools

Collection of packages to do common stuff for the RoboCup SSL league like reading, writing, sending, receiving and parsing messages.

# Install

Use go get to install the packages / executables you like:

```
# analyze match durations
go get github.com/RoboCup-SSL/ssl-go-tools/matchduration

# analyze number of cards given to teams
go get github.com/RoboCup-SSL/ssl-go-tools/numcards

...
```

# Run
The executables are installed to your $GOPATH/bin folder. If you have it on your $PATH, you can directly run them. Else, switch to this folder first.

```
matchduration <directory containing log files>
numcards <directory containing log files>
```
Log files may be compressed, if they end with `.gz`
