# zmetrics

Software to query a zcashd rpc interface for information about blocks and write the data to a json file.

## Usage

```
 ./zmetrics --help
zmetrics gathers data about the Zcash blockchain

Usage:
  zmetrics [flags]

Flags:
      --config string         config file (default is current directory, zmetric.yaml)
      --end-height int        Ending block height (working backwards)
  -h, --help                  help for zmetrics
      --log-level uint32      log level (logrus 1-7) (default 4)
      --num-blocks int        Number of blocks (default 10)
      --output-dir string     Output directory (default "./blocks")
      --rpc-host string       rpc host (default "127.0.0.1")
      --rpc-password string   rpc password (default "notsecret")
      --rpc-port string       rpc port (default "38232")
      --rpc-user string       rpc user account (default "zcashrpc")
      --start-height int      Starting block height, defaults to current height (working backwards)
```

Example `zmetric.yaml` file

```
num-blocks: 100
output-dir: /var/www/html/data
rpc-host: zcashd01.z.cash
rpc-user: zcashrpc
rpc-password: notverysecure
```

Example run to get the last 30 blocks, write them to the current directory:
```
zmetrics --rpc-host 192.168.86.41 \
  --rpc-port 38232 \
  --rpc-user zcashrpc \
  --rpc-password notsecure \
  --num-blocks 30 \
  --output-dir .
```

Example [data](data/zcashmetrics.json)