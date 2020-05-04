# Elrond deploy go

The go implementation for the Elrond deploy

# Getting started

### Prerequisites

Building the repository requires Go (version 1.12 or later)

### Installation and running

Run in  %project_folder%/cmd/filegen folder the following command to build a filegen (which generates .pem and .json
 files used by the node):
 
 ```
 $ go build
$ ./filegen -node-price 500000000000000000000000 -total-supply 20000000000000000000000000000 -num-of-shards 5 -num-of-nodes-in-each-shard 21 -consensus-group-size 15 -num-of-observers-in-each-shard 1 -num-of-metachain-nodes 21 -metachain-consensus-group-size 15 -num-of-observers-in-metachain 1 -chain-id testnet -hysteresis 0.0
 ```

There is an optional flag called `-stake-type` that can have one of the 2 values: `direct` and `delegated`. If the `delegated` 
option is selected, the initial stking will be done through the delegation SC.
 
### Running the tests
```
$ go test ./...
```
