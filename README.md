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
$ ./filegen -mint-value 1000000000000000000000000000 -num-of-shards 5 -num-of-nodes-in-each-shard 21 -consensus-group-size 15 -num-of-observers-in-each-shard 1 -num-of-metachain-nodes 21 -metachain-consensus-group-size 15 -num-of-observers-in-metachain 1 -consensus-type "bls"
 ```
 
### Running the tests
```
$ go test ./...
```
