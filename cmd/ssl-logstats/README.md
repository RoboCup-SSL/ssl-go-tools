# ssl-logstats

A collection of statistics that are gathered from a given set of [official SSL log files](http://wiki.robocup.org/Small_Size_League/Game_Logs).

Following statistics are currently supported:

 * Vision detection frame timings, like average dt and dts > 80ms.
 * Export vision detection frame timings to a CSV file for further analysis. A Matlab script is provided to plot this data
 
## Installation

Use go-get to install this executable:

```
go get -u github.com/RoboCup-SSL/ssl-go-tools/cmd/ssl-logstats
```

## Usage

The binary is called `ssl-logstats`. Run it with `-h` to get the available parameters. A list of log files that should be processed must be provided after the parameters.
