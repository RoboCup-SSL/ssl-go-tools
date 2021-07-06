# ssl-log-player

Play back official log files as described on the [SSL website](https://ssl.robocup.org/game-logs/) in real time. Both, uncompressed and compressed log files are supported. Compressed log files should end with `.gz`. 

Vision and referee data will be send into the network and can be watched by any vision/referee client, like the graphical client from [ssl-vision](https://github.com/RoboCup-SSL/ssl-vision) or the more recent [ssl-vision-client](https://github.com/RoboCup-SSL/ssl-vision-client).

There is currently no support for seeking through the log file or change the playback speed. Have a look at [ssl-logtools](https://github.com/RoboCup-SSL/ssl-logtools) for a player with more control.

## Installation

Use go-get to install this executable:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/cmd/ssl-log-player
```

## Usage

The binary is called `ssl-log-player`.
Run it with `-h` to print usage information.
