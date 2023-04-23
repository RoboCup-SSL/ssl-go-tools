# ssl-multicast-sources

Detect IPs of multicast message producers for debugging purposes.
By default, it looks for all multicast groups and ports that are used by default in the SSL.

## Usage

The binary is called `ssl-multicast-sources`.
Pass custom multicast IPs/ports like this:

```shell
ssl-multicast-sources 224.5.23.1:10003 224.5.23.2:10006
```
