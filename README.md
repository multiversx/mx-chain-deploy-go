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
$ ./filegen -num-addresses-with-balances 21 -mint-value 1000000000 -num-nodes 21 -num-metachain-nodes 1 -consensus-type "bls"
 ```
 
### Running the tests
```
$ go test ./...
```
