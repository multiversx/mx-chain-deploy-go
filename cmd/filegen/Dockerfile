
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
