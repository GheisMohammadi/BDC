# [![bdc](/assets/bdc.png)](https://github.com/GheisMohammadi/BDC) bdc
A simple and integrity PoW blockchain implementation in Golang using ipfs/libp2p 
# install
`
go mod init
go mod tidy
go mod vendor
go build
`
notice: there are some major issues with compiling of go-ipfs. Make sure it can be built properly. 


# install logger
BDC uses viper as logging service.
## install viper
`
env GO111MODULE=on go get github.com/spf13/viper
`
## test
to test all packages use command below

`
go test ./src/...
`
