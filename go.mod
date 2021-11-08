module github.com/ElrondNetwork/elrond-deploy-go

go 1.13

require (
	github.com/ElrondNetwork/elrond-go v1.2.32
	github.com/ElrondNetwork/elrond-go-core v1.1.2
	github.com/ElrondNetwork/elrond-go-crypto v1.0.1
	github.com/ElrondNetwork/elrond-go-logger v1.0.5
	github.com/ElrondNetwork/elrond-vm-common v1.2.4
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.30 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.30

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.30 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.30

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.22 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.22
