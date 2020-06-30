
FROM golang:1.13.6 as builder
MAINTAINER ElrondNetwork

WORKDIR /go/elrond-deploy-go
COPY . .
WORKDIR /go/elrond-deploy-go/cmd/filegen
RUN GO111MODULE=on go mod vendor
RUN go build

# Second stage
FROM ubuntu:18.04
COPY --from=builder "/go/elrond-deploy-go/cmd/filegen" "/go/elrond-deploy-go/cmd/filegen"
WORKDIR /data
ENTRYPOINT ["/go/elrond-deploy-go/cmd/filegen/filegen"]

#RUN ["./filegen", "-total-supply", "1000000000000000000000000000", "-num-of-shards", "3", "-num-of-nodes-in-each-shard", "12", "-consensus-group-size", "12", "-num-of-observers-in-each-shard", "1", "-num-of-metachain-nodes", "12", "-metachain-consensus-group-size", "12", "-num-of-observers-in-metachain", "1", "-chain-id", "testnet", "-hysteresis", "0.0"]
#RUN cp /go/elrond-deploy-go/cmd/filegen/*.json /go/elrond-go/cmd/node/config/
#RUN cp /go/elrond-deploy-go/cmd/filegen/*.pem /go/elrond-go/cmd/node/config/
