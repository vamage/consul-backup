# `consul-backup` - consul Backup and Restore tool

This will use consul-api (Go library) to recursively backup and restore all your key/value pairs. You need to set up your Go environment, then run `make`, which will generate executable named `consul-backup`.

You need to set up your Go environment and "go get github.com/hashicorp/consul/api"
and "go get github.com/docopt/docopt-go"

a "go build" will generate executable named "consul-backup"

##Usage
```sh
Usage:
  consul-backup [-i IP:PORT] [-t TOKEN] [--kv] [--kvfile KVBACKUPFILE] [--acl] [--aclfile ACLBACKUPFILE] [--restore]
  consul-backup -h | --help
  consul-backup --version
```

##Options
```sh
Options:
  -h --help                          Show this screen.
  --version                          Show version.
  -i, --address=IP:PORT              The HTTP endpoint of Consul [default: 127.0.0.1:8500].
  -t, --token=TOKEN                  An ACL Token with proper permissions in Consul [default: ].
  -k, --kv                           Backup or restore KV
  -a, --acl                          Backup or restore ACL
  -K, --kvfile=KVBACKUPFILE          KV Backup Filename [default: kv.bkp].
  -A, --aclfile=ACLBACKUPFILE        ACL Backup Filename [default: acl.bkp].
  -r, --restore                      Activate restore mode
```
