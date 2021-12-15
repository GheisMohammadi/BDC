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

output should be like this:

```
ok      badcoin/src/block       (cached)
ok      badcoin/src/blockchain  (cached)
ok      badcoin/src/config      (cached)
ok      badcoin/src/helper/address      (cached)
ok      badcoin/src/helper/base58       (cached)
ok      badcoin/src/helper/error        (cached)
ok      badcoin/src/helper/file (cached)
ok      badcoin/src/helper/hash (cached)
ok      badcoin/src/helper/logger       (cached)
ok      badcoin/src/helper/number       (cached)
ok      badcoin/src/helper/uuid (cached)
ok      badcoin/src/mempool     (cached)
ok      badcoin/src/merkle      0.002s
ok      badcoin/src/node        (cached)
ok      badcoin/src/pow (cached)
ok      badcoin/src/server      (cached)
ok      badcoin/src/storage     0.022s
ok      badcoin/src/storage/level       (cached)
ok      badcoin/src/transaction (cached)
ok      badcoin/src/wallet      0.002s
```

and for test coverage use command below

`
go test ./src/... -coverprofile=./c.out && go tool cover -html=c.out && unlink ./c.out
`
