# Elrond deploy go

The go implementation for the Elrond deploy

# Getting started

## Prerequisites

Building the repository requires Go (version 1.12 or later)

## Installation and running

Run in  %project_folder%/cmd/filegen folder the following command to build a filegen (which generates .pem and .json
 files used by the node):
 
 ### To run in "direct" staking mode, please run this:
 ```
 $ go build
$ ./filegen -stake-type direct -node-price 2500000000000000000000 -total-supply 20000000000000000000000000 -num-of-shards 5 -num-of-nodes-in-each-shard 21 -consensus-group-size 15 -num-of-observers-in-each-shard 1 -num-of-metachain-nodes 21 -metachain-consensus-group-size 15 -num-of-observers-in-metachain 1 -chain-id testnet -hysteresis 0.0
 ```

 ### To run in "delegated" staking mode, please run this:
 ```
 $ go build
$ ./filegen -stake-type delegated -node-price 2500000000000000000000 -total-supply 20000000000000000000000000 -num-of-shards 5 -num-of-nodes-in-each-shard 21 -consensus-group-size 15 -num-of-observers-in-each-shard 1 -num-of-metachain-nodes 21 -metachain-consensus-group-size 15 -num-of-observers-in-metachain 1 -chain-id testnet -hysteresis 0.0 -num-delegators 1293
 ```

In the "delegated" mode the  initial staking will be done through the delegation SC. When set to "delegated" mode, there are 
2 optional flags `delegation-init` for setting a custom init string and `delegation-version` for setting the correspondent 
delegation SC version.

### Important: 
If the hysteresis value is greater than 0, the binary will add more nodes as validators in order to 
compensate for the nodes in the waiting list. 

### Notes: 
The optional flag called `-richest-account` can be used in order to increase the first wallet key to almost 
all available balance left after the staking process occurred. This is helpful when dealing with automated staking scenarios.

### Running with docker
```
$ docker pull elrondnetwork/elrond-go-filegen:tagname
$ docker run -v /tmp/:/data/ elrondnetwork/elrond-go-filegen:latest -stake-type direct -node-price 2500000000000000000000 -total-supply 20000000000000000000000000 -num-of-shards 5 ...
```
This will create the files on the host machine running Docker at the path location `/tmp/`.
Detailed information about the build is located under https://hub.docker.com/r/elrondnetwork/elrond-go-filegen
 
## Running the tests
```
$ go test ./...
```
