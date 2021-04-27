# ssl-auto-recorder

Record official log files as described on the [SSL website](https://ssl.robocup.org/game-logs/).
The auto-recorder listens for referee messages and starts and stops recordings based on game stages.

## Installation

Use go-get to install this executable:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/cmd/ssl-auto-recorder
```

## Usage

The binary is called `ssl-auto-recorder`.
Run it with `-h` to print usage information.
Logs will be written to a compressed log file in the current working directory.
