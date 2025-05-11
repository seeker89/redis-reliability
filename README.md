# Redis resiliency toolkit

Understand, validate & demonstrate Redis fault tolerance.

# TL;DR

You're probably using [redis](https://github.com/redis/redis). You probably don't fully understand its failure scenarios.

This repo attemps two things:

1. teach you about Redis failure scenarios
2. provide a tool for implementing failure testing scenarios at home (Chaos Engineering)


**Note**: I have no affiliation with Redis Ltd, I'm merely a concerned citizen seeing bad usage in the wild.

# Table of contents
- [Redis resiliency toolkit](#redis-resiliency-toolkit)
- [TL;DR](#tldr)
- [Table of contents](#table-of-contents)
- [1. Learn about Redis HA](#1-learn-about-redis-ha)
- [2. Use the rtt chaos tool](#2-use-the-rtt-chaos-tool)
  - [Building from sources](#building-from-sources)
  - [Building docker image](#building-docker-image)
  - [General usage](#general-usage)
    - [Subcommands](#subcommands)
    - [Output format](#output-format)
  - [`sentinel` subcommand](#sentinel-subcommand)


# 1. Learn about Redis HA

Read [the tutorial here](./book/)

# 2. Use the rtt chaos tool

## Building from sources

```sh
make bin/rrt
```

```sh
./bin/rrt version
{"build":"Sat May 10 14:01:13 BST 2025","version":"v0.0.1"}
```

## Building docker image

```sh
make image
```

## General usage

### Subcommands

`rrt` is split into subcommands:

- `sentinel`
- `kube`
- `chaos`

More details below.

### Output format

`rrt` can output text (`-o text`), more text (`-o wide`) or JSON (`-o json`), that can be optionally `--pretty`.

For example:

```sh
./bin/rrt \
  -o json --pretty \
  sentinel --sentinel $URL_S \
  status
```

will print something like this:

```sh
{
  "host": "exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local",
  "port": "6379"
}
```

or like this if you skip the `--pretty` part:

```json
{"host":"exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local","port":"6379"}
```

or like this if you swap to `-o text`:

```sh
host                                                                      port  
exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379
```

## `sentinel` subcommand

The sentinel command makes it easy to interact with `redis sentinel`:

```sh
 ./bin/rrt sentinel            
Verify Redis sentinel setup

Usage:
  rrt sentinel [command]

Available Commands:
  failover    Trigger soft redis failover
  master      Show the details of the redis master
  replicas    Show the details of the replicas for a master
  sentinels   Show the sentinels for a master
  status      Show the current master of the cluster
  watch       Watch all events on the sentinel

Flags:
  -h, --help              help for sentinel
      --master string     Redis master name (default "mymaster")
      --sentinel string   Redis URL of the sentinel. Use RRT_SENTINEL_URL (default "redis://127.0.0.1:63055")

Global Flags:
      --kube-config string   Path to a kubeconfig file. Leave empty for in-cluster
  -o, --output string        Output format (json, text, wide) (default "json")
  -p, --pretty               Make the output pretty
  -v, --verbose              Make the output verbose

Use "rrt sentinel [command] --help" for more information about a command.
```

You're going to need to specify the sentinel URL. You can use `--sentinel` flag or the `RRT_SENTINEL_URL` envvar.

