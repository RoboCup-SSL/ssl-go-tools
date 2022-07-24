# ssl-log-converter

Convert log files to a human-readable format - a JSON stream.

## Usage

The binary is called `ssl-log-converter`.
Run it with `-h` to print usage information.

The output will be written to `<input-file>.txt`.

## Use cases

You can convert detection or geometry messages to a text file containing one message per line, encoded as JSON.
Afterwards, you can use common Unix CLI tools and `jq`:
```shell
# Extract detection frames
ssl-log-converter -extractDetection my-logfile.log
# Select two fields, convert them to CSV and write them to a file
jq -r '[.camera_id, .t_capture] | @csv' my-logfile.log > data.csv
```
