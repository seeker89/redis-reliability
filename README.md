# Redis resiliency toolkit

Understand, validate & demonstrate Redis fault tolerance.

# TL;DR

You're probably using [redis](https://github.com/redis/redis). You probably don't fully understand its failure scenarios.

This repo attemps two things:

* teach you about Redis failure scenarios
* provide a tool for implementing failure testing scenarios at home (Chaos Engineering)


**Note**: I have no affiliation with Redis Ltd, I'm merely a concerned citizen seeing bad usage in the wild.

# Table of contents
- [Redis resiliency toolkit](#redis-resiliency-toolkit)
- [TL;DR](#tldr)
- [Table of contents](#table-of-contents)
- [Intro to Redis HA](#intro-to-redis-ha)
  - [Redis availability](#redis-availability)
    - [Redis replication](#redis-replication)
    - [Redis sentinel](#redis-sentinel)
    - [Redis cluster](#redis-cluster)
  - [Redis persistence](#redis-persistence)
  - [What we're going to do](#what-were-going-to-do)
- [Setup](#setup)
  - [Kubernetes](#kubernetes)
  - [Helm chart](#helm-chart)
  - [`redis-cli`](#redis-cli)
  - [`k9s` (optional)](#k9s-optional)
- [Exercise 0: replication mode](#exercise-0-replication-mode)
  - [Step 1: prepare the yamls](#step-1-prepare-the-yamls)
  - [Step 2: apply the yamls](#step-2-apply-the-yamls)
  - [Step 3: validate our cluster](#step-3-validate-our-cluster)
  - [Step 4: write some data](#step-4-write-some-data)
  - [Step 6: programmatic access](#step-6-programmatic-access)
  - [Step 7: fly in the ointment (no strong consistency)](#step-7-fly-in-the-ointment-no-strong-consistency)
  - [Step 8: clean up](#step-8-clean-up)
- [Exercise 1: failover in Sentinel mode](#exercise-1-failover-in-sentinel-mode)
  - [Step 0: deploy a version with Sentinel](#step-0-deploy-a-version-with-sentinel)
  - [Step 1: manual failover](#step-1-manual-failover)


# Intro to Redis HA

Most people know `redis` as a fast, in-memory server to store simple data in, for example a cache. But `redis` offers multiple settings to tweak the persistence and availability model to suit your needs, and can do much more than that.

**Spoiler alert**: `redis` can't offer strong consistency.

**Note**: we're only talking about the free (community) edition of `redis` here.

## Redis availability

To make `redis` HA, we have three mechanisms at play:
- replication
- sentinel mode
- cluster mode

### Redis replication

`redis` comes out of the box with [replication](https://redis.io/docs/latest/operate/oss_and_stack/management/replication/).

In short, it allows you to set a number of instances (`replicas`) that will stream commands from another instance (`master`) to stay up to date asynchronously (eventually consistent). All writes go to the `master`, but data can be read from `replicas` as well (potentially state).

If the `master` goes down, a `replica` can be promoted to be a new `master`, and start receiving writes. The other `replicas` need to be repointed to the new `master`. This is called `failover` and doesn't happen automatically.

**Note**: this does not guarantee strong consistency. Replicas can get behind, and a successful write to a `master` can be lost during the failover process.

[Learn more here](https://redis.io/docs/latest/operate/oss_and_stack/management/replication/).


### Redis sentinel

To address the problem of manual `failover`, `redis` offers an HA mode called [`sentinel`](https://redis.io/docs/latest/operate/oss_and_stack/management/sentinel/).

In short, it adds a new set of `redis` instances, whose job is to detect a failed `master` and promote a `replica` to become a `master`. The `sentinel` instances also provide a service discovery ("where should I write?") that's handled through a compatible client library.

**Note**: this automates the `failover` process, but doesn't change the consistency model of `redis`, and it leaves it to the user to design the architecture.

[Learn more here](https://redis.io/docs/latest/operate/oss_and_stack/management/sentinel/).


### Redis cluster

The third (confusingly named) option available to you is the [`redis cluster`](https://redis.io/docs/latest/operate/oss_and_stack/management/scaling/) mode.

In short, as an alternative to `sentinel`, this mode introduces `sharding`, whereby a key is only written to a single `master` instanced, depending on the hash value of the key.

This allows larger `redis` databases, which wouldn't fit on a single node.

In this model, there are multiple `master` nodes, each followed by a number of `replicas`

**Note**: this still uses the same asynchronous replication, and still doesn't guarantee strong consistency.

[Learn more here](https://redis.io/docs/latest/operate/oss_and_stack/management/scaling/).

## Redis persistence

To mix things up further, `redis` supports various [perisistence](https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/) options.

In short:

* you can write point-in-time snapshots (backups)
* you can write write-ahead-log (for each write, or once a second)
* you can do both
* you can do neither

This lets you achieve the level of persistence that you're comfortable with, varying from none (in memory only), to persist every write.

[Learn more here](https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/).


## What we're going to do

With all of that in mind, the rest of the repo will show you how to validate your `redis` setup to ensure that it's actually HA.

To do that, we'll do 3 exercises:

0. demonstrate the lack of strong consistency in `master`-`replica` replication
1. validate `failover` in `sentinel` mode (1 master lost)


# Setup

In this guide, we're going to assume you're reading this in 2025 and you're going to run `redis` on [Kubernetes](https://kubernetes.io/). If you're not, feel free to send a PR.


## Kubernetes

Get a Kubernetes cluster with `kubectl` access to a namespace. We're going to create some `redis` instances in there. If you're new to this stuff, and want to run it on your laptop, [`minikube`](https://minikube.sigs.k8s.io/docs/) is a good place to start.


## Helm chart

We're going to use the [bitnami redis chart](https://github.com/bitnami/charts/tree/main/bitnami/redis), so you're also going to need [helm](https://helm.sh/docs/helm/helm_install/).


## `redis-cli`

We're also going to use the `redis-cli` to play around with our clusters. [Install it](https://redis.io/docs/latest/operate/oss_and_stack/install/archive/install-redis/).


## `k9s` (optional)

All the cool kids are running [`k9s`](https://k9scli.io/) these days. It can help.


# Exercise 0: replication mode

## Step 1: prepare the yamls

We're going to get the bitnami `helm` chart:

```sh
git clone git@github.com:bitnami/charts.git
```

Let's create a minimalistic values file to configure it:

```sh
cat > 0-replication.yaml <<EOF
architecture: replication
# minimal cluster
master:
    replicaCount: 1
replica:
    replicaCount: 2
# much insecure, very wow
auth:
  enabled: false
EOF
```

Now we should be able to run `helm`. Start by downloading the dependencies:

```sh
helm dependency build charts/bitnami/redis
```

Then run `helm template` to get the resulting yamls:

```sh
helm template \
  exercise0 \
  charts/bitnami/redis \
  -f 0-replication.yaml \
  --namespace default \
  --output-dir 0-yamls
```

**Note**: change the namespace to the desired value.

Don't get too scared to see a whole bunch of things:

```
wrote 0-yamls/redis/templates/networkpolicy.yaml
wrote 0-yamls/redis/templates/master/pdb.yaml
wrote 0-yamls/redis/templates/replicas/pdb.yaml
wrote 0-yamls/redis/templates/master/serviceaccount.yaml
wrote 0-yamls/redis/templates/replicas/serviceaccount.yaml
wrote 0-yamls/redis/templates/configmap.yaml
wrote 0-yamls/redis/templates/health-configmap.yaml
wrote 0-yamls/redis/templates/scripts-configmap.yaml
wrote 0-yamls/redis/templates/headless-svc.yaml
wrote 0-yamls/redis/templates/master/service.yaml
wrote 0-yamls/redis/templates/replicas/service.yaml
wrote 0-yamls/redis/templates/master/application.yaml
wrote 0-yamls/redis/templates/replicas/application.yaml
```

Take a moment to take it all in. This is a "serious" `redis` deployment with a lot of bells and whistles. But at the base of it, it's just two `StatefulSet`s with all the glue that makes modern platform engineer secure in their job!


## Step 2: apply the yamls

On my `minikube` setup I'm testing this on, it's just:

```sh
kubectl apply --recursive -f 0-yamls/
```

... and a lot of things are happening:

```
configmap/exercise0-redis-configuration created
service/exercise0-redis-headless created
configmap/exercise0-redis-health created
statefulset.apps/exercise0-redis-master created
poddisruptionbudget.policy/exercise0-redis-master created
service/exercise0-redis-master created
serviceaccount/exercise0-redis-master created
networkpolicy.networking.k8s.io/exercise0-redis created
statefulset.apps/exercise0-redis-replicas created
poddisruptionbudget.policy/exercise0-redis-replicas created
service/exercise0-redis-replicas created
serviceaccount/exercise0-redis-replica created
configmap/exercise0-redis-scripts created
```

A few moments later, I see my master and my `replicas`:

```sh
$ kubectl get pods
NAME                         READY   STATUS    RESTARTS   AGE
exercise0-redis-master-0     1/1     Running   0          112s
exercise0-redis-replicas-0   1/1     Running   0          112s
exercise0-redis-replicas-1   1/1     Running   0          77s
```

You'll also see the two services we're interesting in, one for master and one for replicas:

```sh
~ % kubectl get svc
NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
exercise0-redis-headless   ClusterIP   None             <none>        6379/TCP   21h
exercise0-redis-master     ClusterIP   10.111.240.225   <none>        6379/TCP   21h
exercise0-redis-replicas   ClusterIP   10.96.219.7      <none>        6379/TCP   21h
...
```

And because it's a `minikube` setup, I've got one more hoop to jump through:

```sh
minikube service exercise0-redis-master
```

Which should display something like this:

```sh
|-----------|------------------------|-------------|--------------|
| NAMESPACE   | NAME                     | TARGET PORT   | URL            |
| ----------- | ------------------------ | ------------- | -------------- |
| default     | exercise0-redis-master   |               | No node port   |
| ----------- | ------------------------ | ------------- | -------------- |
😿  service default/exercise0-redis-master has no node port
❗  Services [default/exercise0-redis-master] have type "ClusterIP" not meant to be exposed, however for local development minikube allows you to access this !
🏃  Starting tunnel for service exercise0-redis-master.
|-----------|------------------------|-------------|------------------------|
| NAMESPACE   | NAME                     | TARGET PORT   | URL                      |
| ----------- | ------------------------ | ------------- | ------------------------ |
| default     | exercise0-redis-master   |               | http://127.0.0.1:49801   |
| ----------- | ------------------------ | ------------- | ------------------------ |
🎉  Opening service default/exercise0-redis-master in default browser...
❗  Because you are using a Docker driver on darwin, the terminal needs to be open to run it.
```

And now I should be able to connect to the master:

```sh
redis-cli -h 127.0.0.1 -p 49801
127.0.0.1:49801> ping
PONG
127.0.0.1:49801>
```

## Step 3: validate our cluster

For `minikube`rs, make sure that you can access both the `master` and `replicas` services:

```sh
minikube service exercise0-redis-master exercise0-redis-replicas 
```

You should see something like this:

```sh
|-----------|------------------------|-------------|--------------|
| NAMESPACE   | NAME                     | TARGET PORT   | URL            |
| ----------- | ------------------------ | ------------- | -------------- |
| default     | exercise0-redis-master   |               | No node port   |
| ----------- | ------------------------ | ------------- | -------------- |
😿  service default/exercise0-redis-master has no node port
|-----------|--------------------------|-------------|--------------|
| NAMESPACE   | NAME                       | TARGET PORT   | URL            |
| ----------- | -------------------------- | ------------- | -------------- |
| default     | exercise0-redis-replicas   |               | No node port   |
| ----------- | -------------------------- | ------------- | -------------- |
😿  service default/exercise0-redis-replicas has no node port
❗  Services [default/exercise0-redis-master default/exercise0-redis-replicas] have type "ClusterIP" not meant to be exposed, however for local development minikube allows you to access this !
🏃  Starting tunnel for service exercise0-redis-master.
🏃  Starting tunnel for service exercise0-redis-replicas.
|-----------|--------------------------|-------------|------------------------|
| NAMESPACE   | NAME                       | TARGET PORT   | URL                      |
| ----------- | -------------------------- | ------------- | ------------------------ |
| default     | exercise0-redis-master     |               | http://127.0.0.1:63057   |
| default     | exercise0-redis-replicas   |               | http://127.0.0.1:63059   |
| ----------- | -------------------------- | ------------- | ------------------------ |
...
```

It's complaining about `ClusterIP`, but then it's happily obliging, so we'll leave it like that.

For convenience, let's set some variables so that we can call them easily:

```sh
export URL_M=redis://127.0.0.1:63057
export URL_R=redis://127.0.0.1:63059
```

And let's say hi to see all's good:

```sh
~ % redis-cli -u $URL_M HELLO
 1) "server"
 2) "redis"
 3) "version"
 4) "7.4.2"
 5) "proto"
 6) (integer) 2
 7) "id"
 8) (integer) 93943
 9) "mode"
10) "standalone"
11) "role"
12) "master"
13) "modules"
14) (empty array)
```

```sh
~ % redis-cli -u $URL_R HELLO
 1) "server"
 2) "redis"
 3) "version"
 4) "7.4.2"
 5) "proto"
 6) (integer) 2
 7) "id"
 8) (integer) 31325
 9) "mode"
10) "standalone"
11) "role"
12) "replica"
13) "modules"
14) (empty array)
```

You can also ask specifically about the replication status.

The `master` node will tell you about the `replicas`:

```sh
% redis-cli -u $URL_M INFO replication
# Replication
role:master
connected_slaves:2
slave0:ip=exercise0-redis-replicas-0.exercise0-redis-headless.default.svc.cluster.local,port=6379,state=online,offset=107913,lag=0
slave1:ip=exercise0-redis-replicas-1.exercise0-redis-headless.default.svc.cluster.local,port=6379,state=online,offset=107913,lag=0
master_failover_state:no-failover
master_replid:db4ef924de6379dd2150a807925ac267de278ef0
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:107913
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset:1
repl_backlog_histlen:107913
```

And the `replicas` will point to the `master`:

```sh
% redis-cli -u $URL_R INFO replication
# Replication
role:slave
master_host:exercise0-redis-master-0.exercise0-redis-headless.default.svc.cluster.local
master_port:6379
master_link_status:up
master_last_io_seconds_ago:7
master_sync_in_progress:0
slave_read_repl_offset:107913
slave_repl_offset:107913
slave_priority:100
slave_read_only:1
replica_announced:1
connected_slaves:0
master_failover_state:no-failover
master_replid:db4ef924de6379dd2150a807925ac267de278ef0
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:107913
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset:1
repl_backlog_histlen:107913
```

**Note**: we have two `replicas` behind that single service, so you're hitting one at a time.


## Step 4: write some data

Now we're ready to see the data being replicated. How exciting!

Let's first see I'm not cheating:

```sh
% redis-cli -u $URL_R GET mystery                                            
(nil)
% redis-cli -u $URL_M GET mystery
(nil)
```

Now let's confirm the replica is read only:

```sh
% redis-cli -u $URL_R SET mystery something
(error) READONLY You can't write against a read only replica.
```

Finally, let's write some important data:

```sh
% redis-cli -u $URL_M SET mystery https://www.youtube.com/watch?v=dQw4w9WgXcQ
OK
```

And let's confirm that both the replicas and master have the data:

```sh
% redis-cli -u $URL_R GET mystery                                            
"https://www.youtube.com/watch?v=dQw4w9WgXcQ"
% redis-cli -u $URL_M GET mystery                                            
"https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

Go on, click it. I know you want to.

## Step 6: programmatic access

So the cli is all nice and good, but we're going to need something tad more automated.

[cmd/hello-redis](./cmd/hello-redis/main.go) is the hello world of connecting to `redis` with go.

Let's run that to confirm everything's in order:

```sh
go run cmd/hello-redis/main.go
```

You should see our mystery string again, read from that `URL_M` variable:

```sh
connecting to redis: 127.0.0.1:63057
mystery: https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

Cool, so now we have everything we need to test out that consistency.


## Step 7: fly in the ointment (no strong consistency)

Our little adventure so far might give you a false sense of security. Let's rain on that parade!

To show the stale reads from replicas, we're going to simulate 100 clients, each making a 1000 operations.

We'll start `writer` goroutines which will write to a key, and corresponding `reader` goroutines that will read it as soon as the writer got acknoledgement, and compare the values.

My choice of Go to write the examples should now become clearer.

[cmd/async-replication](./cmd/async-replication/main.go) is where the code lives. All 90ish lines of it.


```sh
go run cmd/async-replication/main.go
```

You should see an output similar to this one, with random order:


```sh
...
Wrong value: client_87 got: 826 expected: 827
Wrong value: client_80 got: 802 expected: 803
Wrong value: client_81 got: 799 expected: 800
Done: client_97 total_reads: 1000 stale_reads: 3 error_rate: 0.003
Done: client_70 total_reads: 1000 stale_reads: 1 error_rate: 0.001
Done: client_67 total_reads: 1000 stale_reads: 0 error_rate: 0
Done: client_92 total_reads: 1000 stale_reads: 1 error_rate: 0.001
Done: client_22 total_reads: 1000 stale_reads: 4 error_rate: 0.004
Done: client_65 total_reads: 1000 stale_reads: 1 error_rate: 0.001
Done: client_99 total_reads: 1000 stale_reads: 2 error_rate: 0.002
Done: client_24 total_reads: 1000 stale_reads: 3 error_rate: 0.003
Done: client_54 total_reads: 1000 stale_reads: 3 error_rate: 0.003
Done: client_82 total_reads: 1000 stale_reads: 2 error_rate: 0.002
Done: client_55 total_reads: 1000 stale_reads: 2 error_rate: 0.002
Done: client_73 total_reads: 1000 stale_reads: 3 error_rate: 0.003
Done: client_56 total_reads: 1000 stale_reads: 2 error_rate: 0.002
Done: client_96 total_reads: 1000 stale_reads: 2 error_rate: 0.002
Done: client_71 total_reads: 1000 stale_reads: 3 error_rate: 0.003
Done: client_86 total_reads: 1000 stale_reads: 2 error_rate: 0.002
...
100 clients all done
```

## Step 8: clean up 

Let's wind everything down:

```sh
kubectl delete --recursive -f 0-yamls
```

And we can clean up the env vars we set:

```sh
unset URL_M
unset URL_R
```


# Exercise 1: failover in Sentinel mode

Let's now take a look at what the situation looks like when you throw in a `sentinel` into the mix.

## Step 0: deploy a version with Sentinel

I prepped a new file [1-sentinel.yaml](./1-sentinel.yaml) with a minimal config to give us a sentinel-enabled cluster.

Let's see what it does:

```sh
helm template \
  exercise1 \
  charts/bitnami/redis \
  -f 1-sentinel.yaml \
  --namespace default \
  --output-dir 1-yamls
```

Surprise - this time we actually get fewer things:

```sh
wrote 1-yamls/redis/templates/networkpolicy.yaml
wrote 1-yamls/redis/templates/sentinel/pdb.yaml
wrote 1-yamls/redis/templates/serviceaccount.yaml
wrote 1-yamls/redis/templates/configmap.yaml
wrote 1-yamls/redis/templates/health-configmap.yaml
wrote 1-yamls/redis/templates/scripts-configmap.yaml
wrote 1-yamls/redis/templates/headless-svc.yaml
wrote 1-yamls/redis/templates/sentinel/service.yaml
wrote 1-yamls/redis/templates/sentinel/statefulset.yaml
```

We only get one stateful set, and it contains two containers: one for redis, and one for the sentinel.

Let's give it a whirl:

```sh
kubectl apply --recursive -f 1-yamls/
```

After a little while, we'll see our 3 replicas:

```sh
% kubectl get pods
NAME                     READY   STATUS    RESTARTS   AGE
exercise1-redis-node-0   2/2     Running   0          2m5s
exercise1-redis-node-1   2/2     Running   0          91s
exercise1-redis-node-2   2/2     Running   0          68s
```

**Note**: the 2/2 showing both the redis and sentinel containers.


This time we don't get two services for writing and reading. Instead, we get a single service (in a regular and headless variant) with two ports: one for redis (for reading) and one for sentinel (for service discovery).

```sh
% kubectl get svc 
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)              AGE
exercise1-redis            ClusterIP   10.100.110.77   <none>        6379/TCP,26379/TCP   3m36s
exercise1-redis-headless   ClusterIP   None            <none>        6379/TCP,26379/TCP   3m36s
kubernetes                 ClusterIP   10.96.0.1       <none>        443/TCP              3d5h
```

Once again, for those on `minikube`, expose the service:

```sh
minikube service exercise1-redis
```

For some reason it doesn't display the target port in the table, so we'll have to check which one is which:

```sh
...
|-----------|-----------------|-------------|------------------------|
| NAMESPACE   | NAME              | TARGET PORT   | URL                      |
| ----------- | ----------------- | ------------- | ------------------------ |
| default     | exercise1-redis   |               | http://127.0.0.1:63054   |
|             |                   |               | http://127.0.0.1:63055   |
| ----------- | ----------------- | ------------- | ------------------------ |
...
```

They happen to be in the same order as yaml, so let's export these:

```sh
export URL_R=redis://127.0.0.1:63054
export URL_S=redis://127.0.0.1:63055
```

Note, that `URL_R` now points to a random node, one of which is a currently a master. If you call the familiar `INFO replication` command on that URL a few times, you should hit all the nodes.

This is what my `master` looks like:

```sh
% redis-cli -u $URL_R INFO replication
# Replication
role:master
connected_slaves:2
slave0:ip=exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local,port=6379,state=online,offset=975285,lag=1
slave1:ip=exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local,port=6379,state=online,offset=975547,lag=0
master_failover_state:no-failover
master_replid:f520c2fee645b2e7a9966ac5b3d04ca53e15ea30
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:975547
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset:1
repl_backlog_histlen:975547
```

And this is a `replica`:

```sh
% redis-cli -u $URL_R INFO replication
# Replication
role:slave
master_host:exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local
master_port:6379
master_link_status:up
master_last_io_seconds_ago:0
master_sync_in_progress:0
slave_read_repl_offset:981849
slave_repl_offset:981849
slave_priority:100
slave_read_only:1
replica_announced:1
connected_slaves:0
master_failover_state:no-failover
master_replid:f520c2fee645b2e7a9966ac5b3d04ca53e15ea30
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:981849
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset:7687
repl_backlog_histlen:974163
```

To get a definitive answer, we can call the `sentinel` and use the [`SENTINEL` command](https://cndoc.github.io/redis-doc-cn/cn/topics/sentinel.html#sentinel-api) to query for a master.

It turns out to be the catchy `get-master-addr-by-name` subcommand:

```sh
% redis-cli -u $URL_S SENTINEL get-master-addr-by-name mymaster
1) "exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local"
2) "6379"
```

## Step 1: manual failover

Let's see what happens when we manually tell the sentinel to failover the master.

```sh
% redis-cli -u $URL_S SENTINEL failover mymaster
OK
```

Easy peasy:

```sh
% redis-cli -u $URL_S SENTINEL get-master-addr-by-name mymaster
1) "exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local"
2) "6379"
```

What does it look like on the old master? From the logs:

```sh
│ 1:M 25 Apr 2025 20:50:13.114 * Connection with replica exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379 lost.                                                                                                         │
│ 1:M 25 Apr 2025 20:50:14.056 * Connection with replica exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local:6379 lost.                                                                                                         │
│ 1:S 25 Apr 2025 20:50:24.180 * Before turning into a replica, using my own master parameters to synthesize a cached master: I may be able to synchronize with the new master with just a partial transfer.                                          │
│ 1:S 25 Apr 2025 20:50:24.180 * Connecting to MASTER exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379                                                                                                                  │
│ 1:S 25 Apr 2025 20:50:24.181 * MASTER <-> REPLICA sync started                                                                                                                                                                                      │
│ 1:S 25 Apr 2025 20:50:24.181 * REPLICAOF exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379 enabled (user request from 'id=5425 addr=10.244.0.8:55734 laddr=10.244.0.6:6379 fd=13 name=sentinel-b0fe4c03-cmd age=10 idl │
│ 1:S 25 Apr 2025 20:50:24.181 * Non blocking connect for SYNC fired the event.                                                                                                                                                                       │
│ 1:S 25 Apr 2025 20:50:24.181 * Master replied to PING, replication can continue...                                                                                                                                                                  │
│ 1:S 25 Apr 2025 20:50:24.181 * Trying a partial resynchronization (request f520c2fee645b2e7a9966ac5b3d04ca53e15ea30:5226722).                                                                                                                       │
│ 1:S 25 Apr 2025 20:50:29.065 * Full resync from master: 81f947776bb833b3608d5786844b3ab7cb37dd95:5229378                                                                                                                                            │
│ 1:S 25 Apr 2025 20:50:29.071 * MASTER <-> REPLICA sync: receiving streamed RDB from master with EOF to disk                                                                                                                                         │
│ 1:S 25 Apr 2025 20:50:29.072 * Discarding previously cached master state.                                                                                                                                                                           │
│ 1:S 25 Apr 2025 20:50:29.072 * MASTER <-> REPLICA sync: Flushing old data                                                                                                                                                                           │
│ 1:S 25 Apr 2025 20:50:29.073 * MASTER <-> REPLICA sync: Loading DB in memory 
...
```

And this is what the new `master` has to say:

```sh
│ 1:M 25 Apr 2025 20:50:13.114 * Connection with master lost.                                                                                                                                                                                         │
│ 1:M 25 Apr 2025 20:50:13.114 * Caching the disconnected master state.                                                                                                                                                                               │
│ 1:M 25 Apr 2025 20:50:13.114 * Discarding previously cached master state.                                                                                                                                                                           │
│ 1:M 25 Apr 2025 20:50:13.114 * Setting secondary replication ID to f520c2fee645b2e7a9966ac5b3d04ca53e15ea30, valid up to offset: 5223316. New replication ID is 81f947776bb833b3608d5786844b3ab7cb37dd95                                            │
│ 1:M 25 Apr 2025 20:50:13.115 * MASTER MODE enabled (user request from 'id=3 addr=10.244.0.7:39464 laddr=10.244.0.7:6379 fd=14 name=sentinel-b460d64b-cmd age=13516 idle=0 flags=x db=0 sub=0 psub=0 ssub=0 multi=4 watch=0 qbuf=188 qbuf-free=20286 │
│ 1:M 25 Apr 2025 20:50:14.059 * Replica exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local:6379 asks for synchronization                                                                                                      │
│ 1:M 25 Apr 2025 20:50:14.059 * Partial resynchronization request from exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local:6379 accepted. Sending 285 bytes of backlog starting from offset 5223316.                           │
│ 1:M 25 Apr 2025 20:50:24.182 * Replica exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local:6379 asks for synchronization                                                                                                      │
│ 1:M 25 Apr 2025 20:50:24.182 * Partial resynchronization not accepted: Requested offset for second ID was 5226722, but I can reply up to 5223316                                                                                                    │
│ 1:M 25 Apr 2025 20:50:24.182 * Delay next BGSAVE for diskless SYNC                                                                                                                                                                                  │
│ 1:M 25 Apr 2025 20:50:29.065 * Starting BGSAVE for SYNC with target: replicas sockets                                                                                                                                                               │
│ 1:M 25 Apr 2025 20:50:29.067 * Background RDB transfer started by pid 59081                                                                                                                                                                         │
│ 59081:C 25 Apr 2025 20:50:29.070 * Fork CoW for RDB: current 0 MB, peak 0 MB, average 0 MB                                                                                                                                                          │
│ 1:M 25 Apr 2025 20:50:29.070 * Diskless rdb transfer, done reading from pipe, 1 replicas still up.                                                                                                                                                  │
│ 1:M 25 Apr 2025 20:50:29.078 * Background RDB transfer terminated with success                                                                                                                                                                      │
│ 1:M 25 Apr 2025 20:50:29.078 * Streamed RDB transfer with replica exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local:6379 succeeded (socket). Waiting for REPLCONF ACK from replica to enable streaming                      │
│ 1:M 25 Apr 2025 20:50:29.078 * Synchronization with replica exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local:6379 succeeded
```

And finally, the third node, merely switching masters:

```sh
│ 1:M 25 Apr 2025 20:50:14.054 * Connection with master lost.                                                                                                                                                                                         │
│ 1:M 25 Apr 2025 20:50:14.054 * Caching the disconnected master state.                                                                                                                                                                               │
│ 1:S 25 Apr 2025 20:50:14.054 * Connecting to MASTER exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379                                                                                                                  │
│ 1:S 25 Apr 2025 20:50:14.056 * MASTER <-> REPLICA sync started                                                                                                                                                                                      │
│ 1:S 25 Apr 2025 20:50:14.056 * REPLICAOF exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379 enabled (user request from 'id=9 addr=10.244.0.7:45106 laddr=10.244.0.8:6379 fd=12 name=sentinel-b460d64b-cmd age=13486 idl │
│ 1:S 25 Apr 2025 20:50:14.057 * Non blocking connect for SYNC fired the event.                                                                                                                                                                       │
│ 1:S 25 Apr 2025 20:50:14.058 * Master replied to PING, replication can continue...                                                                                                                                                                  │
│ 1:S 25 Apr 2025 20:50:14.058 * Trying a partial resynchronization (request f520c2fee645b2e7a9966ac5b3d04ca53e15ea30:5223316).                                                                                                                       │
│ 1:S 25 Apr 2025 20:50:14.059 * Successful partial resynchronization with master.                                                                                                                                                                    │
│ 1:S 25 Apr 2025 20:50:14.059 * Master replication ID changed to 81f947776bb833b3608d5786844b3ab7cb37dd95                                                                                                                                            │
│ 1:S 25 Apr 2025 20:50:14.059 * MASTER <-> REPLICA sync: Master accepted a Partial Resynchronization.   
```

The `sentinel` instance logged this:

```sh
│ 1:X 25 Apr 2025 20:50:12.889 * Executing user requested FAILOVER of 'mymaster'                                                                                                                                                                      │
│ 1:X 25 Apr 2025 20:50:12.889 # +new-epoch 1                                                                                                                                                                                                         │
│ 1:X 25 Apr 2025 20:50:12.889 # +try-failover master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379                                                                                                         │
│ 1:X 25 Apr 2025 20:50:12.948 * Sentinel new configuration saved on disk                                                                                                                                                                             │
│ 1:X 25 Apr 2025 20:50:12.948 # +vote-for-leader b460d64bf5a7297a9147de030fd36155499923a8 1                                                                                                                                                          │
│ 1:X 25 Apr 2025 20:50:12.948 # +elected-leader master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379                                                                                                       │
│ 1:X 25 Apr 2025 20:50:12.948 # +failover-state-select-slave master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379                                                                                          │
│ 1:X 25 Apr 2025 20:50:13.040 # +selected-slave slave exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exercise1-redis-node- │
│ 1:X 25 Apr 2025 20:50:13.040 * +failover-state-send-slaveof-noone slave exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster ex │
│ 1:X 25 Apr 2025 20:50:13.113 * +failover-state-wait-promotion slave exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exerci │
│ 1:X 25 Apr 2025 20:50:13.977 * Sentinel new configuration saved on disk                                                                                                                                                                             │
│ 1:X 25 Apr 2025 20:50:13.978 # +promoted-slave slave exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exercise1-redis-node- │
│ 1:X 25 Apr 2025 20:50:13.978 # +failover-state-reconf-slaves master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379                                                                                         │
│ 1:X 25 Apr 2025 20:50:14.053 * +slave-reconf-sent slave exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exercise1-redis-no │
│ 1:X 25 Apr 2025 20:50:15.025 * +slave-reconf-inprog slave exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exercise1-redis- │
│ 1:X 25 Apr 2025 20:50:15.025 * +slave-reconf-done slave exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exercise1-redis-no │
│ 1:X 25 Apr 2025 20:50:15.098 # +failover-end master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379                                                                                                         │
│ 1:X 25 Apr 2025 20:50:15.098 # +switch-master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379 exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local 6379                                │
│ 1:X 25 Apr 2025 20:50:15.099 * +slave slave exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-2.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exercise1-redis-node-1.exercis │
│ 1:X 25 Apr 2025 20:50:15.099 * +slave slave exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local:6379 exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379 @ mymaster exercise1-redis-node-1.exercis │
│ 1:X 25 Apr 2025 20:50:15.103 * Sentinel new configuration saved on disk
```

***Note**: the key message here was `+switch-master mymaster exercise1-redis-node-0.exercise1-redis-headless.default.svc.cluster.local 6379 exercise1-redis-node-1.exercise1-redis-headless.default.svc.cluster.local 6379`.
This is probably what you're going to look for in logs while debugging.

[Learn more here](https://cndoc.github.io/redis-doc-cn/cn/topics/sentinel.html).

***Note***: there wasn't a real election for the new master, as this is a user-requested failover. So the quorum wasn't used here.

**Note**: also bear in mind that the failover took about 5 seconds, with minimal data to be synchronised:

```sh
% redis-cli -u $URL_R INFO memory 
# Memory
used_memory:1770168
used_memory_human:1.69M
used_memory_rss:18042880
used_memory_rss_human:17.21M
used_memory_peak:2577096
used_memory_peak_human:2.46M
used_memory_peak_perc:68.69%
used_memory_overhead:1405252
used_memory_startup:979032
used_memory_dataset:364916
used_memory_dataset_perc:46.13%
```

