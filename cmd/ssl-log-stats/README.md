# ssl-log-stats

A collection of statistics that are gathered from a given set of [official SSL log files](https://ssl.robocup.org/game-logs/).

Following statistics are currently supported:

 * Vision detection frame timings, like average dt and dts > 80ms.
 * Export vision detection frame timings to a CSV file for further analysis. A Matlab script is provided to plot this data

## Usage

The binary is called `ssl-log-stats`.
Run it with `-h` to print usage information.
