# ssl-logplayer

Play back official log files as described in the [SSL wiki](http://wiki.robocup.org/Small_Size_League/Game_Logs) in real time. Both, uncompressed and compressed log files are supported. Compressed log files should end with `.gz`. 

Vision and referee data will be send into the network and can be watched by any vision/referee client, like the graphical client from [ssl-vision](https://github.com/RoboCup-SSL/ssl-vision).

There is currently no support for seeking through the log file or change the playback speed. Have a look at [ssl-logtools](https://github.com/RoboCup-SSL/ssl-logtools) for a more player with more control.

## Installation

Use go-get to install this executable:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/ssl-logplayer
```

## Usage

The binary is called `ssl-logplayer`. Run it with `-h` to get the available parameters.