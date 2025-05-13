# Redis resiliency toolkit

Understand, validate & demonstrate Redis fault tolerance.

# TL;DR

You're probably using [`redis`](https://github.com/redis/redis).

You probably don't fully understand its failure scenarios.

This repo attemps to:

1. teach you `redis` failure scenarios
2. give you a tool for implementing failure testing scenarios at home (Chaos Engineering)

**Note**: I have no affiliation with Redis Ltd

# Table of contents
- [Redis resiliency toolkit](#redis-resiliency-toolkit)
- [TL;DR](#tldr)
- [Table of contents](#table-of-contents)
- [1. Learn Redis HA](#1-learn-redis-ha)
- [2. Use `rrt` for resiliency testing](#2-use-rrt-for-resiliency-testing)
  - [Building from sources](#building-from-sources)
  - [Building docker image](#building-docker-image)
  - [General usage](#general-usage)
    - [Subcommands](#subcommands)
    - [Output format](#output-format)
  - [`sentinel` subcommand](#sentinel-subcommand)
    - [`sentinel failover`](#sentinel-failover)
    - [`sentinel kill`](#sentinel-kill)
    - [`sentinel wait`](#sentinel-wait)


# 1. Learn Redis HA

Follow [the tutorial here](./book/) to a self-paced workshop on redis high availability.

# 2. Use `rrt` for resiliency testing

`rrt` is a command line utility designed to make testing `redis` super simple.

It can be used to observe the status of the cluster, plug into other tools & automation (it spits out JSON), and to implement automatic failure injection aka Chaos Engineering. I [like Chaos Engineering](https://www.manning.com/books/chaos-engineering).
 
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

```json
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
  kill        Kill the master to trigger failover
  master      Show the details of the redis master
  replicas    Show the details of the replicas for a master
  sentinels   Show the sentinels for a master
  status      Show the current master of the cluster
  wait        Wait for the new master election
  watch       Watch all events on the sentinel

Flags:
  -g, --grace duration     Grace period for killing
  -h, --help               help for sentinel
      --master string      Redis master name (default "mymaster")
      --sentinel string    Redis URL of the sentinel. Use RRT_SENTINEL_URL (default "redis://127.0.0.1:63055")
  -t, --timeout duration   Timeout for killing (default 1m0s)

Global Flags:
      --kubeconfig string   Path to a kubeconfig file. Leave empty for in-cluster. (KUBECONFIG)
      --namespace string    Limit Kubernetes actions to only this namespace (NAMESPACE)
  -o, --output string       Output format (json, text, wide) (default "json")
  -p, --pretty              Make the output pretty
  -v, --verbose             Make the output verbose

Use "rrt sentinel [command] --help" for more information about a command.
```

You're going to need to specify the sentinel URL. You can use `--sentinel` flag or the `RRT_SENTINEL_URL` envvar.


### `sentinel failover`

Triggers an immediate failover. This is a built-in feature of `redis`. It doesn't wait for any timeouts, doesn't consult the other sentinel instances, and goes and directly elects a new master.

It goes something like this. First, check the current master:

```sh
export RRT_SENTINEL_URL=redis://127.0.0.1:63055
```

```sh
./bin/rrt \
  -o json --pretty \
  sentinel status
```

You will see something like this:

```json
{
  "host": "exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local",
  "port": "6379"
}
```

Next, start a watch so that we can see everything that happens. Run this in one terminal session:

```sh
./bin/rrt \
  -o json --pretty \
  sentinel watch
```

You won't see anything yet.

Now, trigger a soft failover using the `sentinel failover` subcommand in a second terminal:

```sh
./bin/rrt \
  sentinel failover
```

In the first terminal, you will see a bunch of events:

```json
{
  "ch": "+new-epoch",
  "msg": "19",
  "time": "2025-05-06 01:06:45.490267 +0100 BST m=+7.552460126"
}
{
  "ch": "+try-failover",
  "msg": "master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379",
  "time": "2025-05-06 01:06:45.490588 +0100 BST m=+7.552781501"
}
{
  "ch": "+vote-for-leader",
  "msg": "0a1c2be9e2920281bb1de0d299e4acc9c11fea59 19",
  "time": "2025-05-06 01:06:45.513966 +0100 BST m=+7.576159501"
}
{
  "ch": "+elected-leader",
  "msg": "master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379",
  "time": "2025-05-06 01:06:45.514018 +0100 BST m=+7.576211042"
}

...

{
  "ch": "+switch-master",
  "msg": "mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379 exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local 6379",
  "time": "2025-05-06 01:06:51.797919 +0100 BST m=+13.860090917"
}

...
```

And in the second one, you can check the new `master` again:

```json
{
  "host": "exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local",
  "port": "6379"
}
```

### `sentinel kill`

A more drastic (and realistic) version of `sentinel failover`.

It kills the pod returned by the sentinel until any of the following are true:
* a new master is elected
* a timeout has elapsed

:warning: note, that the name of the pod to kill is expected to be stable (`statefulset`) and returned as the first part of the DNS name returned by the sentinel.

You will need access to `kubernetes`, which you can set up by either:

* populating `KUBECONFIG` or `--kubeconfig` with a path to valid `kubectl` config - for running out of the cluster
* setting RBAC on the service account in use - for running in cluster

With that, it's as simple as running `rrt sentinel kill`. For example:

```sh
./bin/rrt \
  sentinel kill \
  --pretty --kubeconfig ~/.kube/config --grace 1s --timeout 5m
```

Will give you something like this:

```sh
{
  "event": "initial master",
  "msg": "exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local:6379",
  "time": "2025-05-13 01:26:25.061459 +0100 BST m=+0.049989542"
}
{
  "event": "deleting pod",
  "name": "exercise1-redis-node-0",
  "namespace": "default",
  "time": "2025-05-13 01:26:25.088399 +0100 BST m=+0.076929376"
}
{
  "event": "deleting pod",
  "name": "exercise1-redis-node-0",
  "namespace": "default",
  "time": "2025-05-13 01:26:25.093569 +0100 BST m=+0.082099959"
}
{
  "event": "done",
  "new_master": "exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379",
  "result": "OK",
  "time": "2025-05-13 01:26:26.366692 +0100 BST m=+1.355222876"
}
{
  "done": "true",
  "event": "final master",
  "msg": "exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379",
  "time": "2025-05-13 01:26:26.367382 +0100 BST m=+1.355912584"
}
```


### `sentinel wait`

Conversely, it's often handy to just wait until a new master is elected.

You can do that easily, like so:

```sh
./bin/rrt sentinel wait --pretty                                            
```

It will just wait there until a master is elected, and then exit and print the diff:

```json
{
  "host": "exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local",
  "port": "6379",
  "previous_host": "exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local",
  "previous_port": "6379"
}
```
