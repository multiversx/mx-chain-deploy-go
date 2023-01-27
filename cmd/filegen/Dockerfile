
FROM golang:1.13.6 as builder
MAINTAINER MultiversX

WORKDIR /go/mx-chain-deploy-go
COPY . .
WORKDIR /go/mx-chain-deploy-go/cmd/filegen
RUN GO111MODULE=on go mod vendor
RUN go build

# Second stage
FROM ubuntu:18.04
COPY --from=builder "/go/mx-chain-deploy-go/cmd/filegen" "/go/mx-chain-deploy-go/cmd/filegen"
WORKDIR /data
ENTRYPOINT ["/go/mx-chain-deploy-go/cmd/filegen/filegen"]
