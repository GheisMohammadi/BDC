# [![bdc](/assets/bdc.png)](https://github.com/GheisMohammadi/BDC) bdc
A simple and integrity PoW blockchain implementation in Golang using ipfs/libp2p 

> A blockchain – originally block chain – is a distributed database that maintains a continuously growing list of ordered records called blocks. Each block contains a timestamp and a link to a previous block. By design, blockchains are inherently resistant to modification of the data — once recorded, the data in a block cannot be altered retroactively. Blockchains are "an open, distributed ledger that can record transactions between two parties efficiently and in a verifiable and permanent way. The ledger itself can also be programmed to trigger transactions automatically." [Wikipedia](https://en.wikipedia.org/wiki/Blockchain_(database))


- [Description](#description)
- [Project Requirements](#project-requirements)
- [Keys](#keys)
- [Building the source](#building-the-source)
- [Test](#test)
- [P2P Networking](#p2p-networking)
- [Block Storage](#block-storage)
- [Mining](#mining)
- [Block Structure](#block-structure)
- [Transaction](#transaction)
- [Wallet](#wallet)
- [CLI](#cli)
- [Server Endpoints](#server-endpoints)

# Description
This is a sample blockchain project which supports a new AltCoin - BadCoin (BDC). 
The chain can be running on two (can be local) or more nodes. BDC uses proof of work as consensus.

# Project Requirements
* The chain needs to be open for anyone to connect to it.
* The target block time is 30 seconds.
* The mining reward is 100 coins per block and halves every 100 blocks.
* The genesis block needs to have a nonce of "1337".
* The blockchain needs to support coin transactions
* The blockchain needs to protect against double spending
* The blockchain needs to allow adding optional extra data to each transaction, which is included in the tx hash
* The blockchain needs to allow adding optional extra data to each block (and there needs to be a way for a miner to set it)
# Keys
The list below are the key features already implemented:
* Core block struct 
* NewBlock 
* Merkle Tree
* GenesisBlock
* Transaction
* Wallet
* Leveldb
* Balance
* MemPool
* DoubleSpend
* ProofOfWork 
* calculateHash
* block reward
* JsonRPC server
* net sync
* protocol (done: version, getaddr, addr, getblocks, getdate, block, tx)
* CLI tools



# Building the source
to build the source, clone the codes on local and open the project directory and run these:

```
$ go mod init

$ go mod tidy

$ go mod vendor

$ go build
```

notice: there are some major issues with compiling of go-ipfs. Make sure it can be built properly. 
## installing libp2p
The go-libp2p library has 2 drawbacks:

1- Setup is not that much easy. It uses gx as a package manager which is not very convenient and still needs improvements.

2- It appears to still be under heavy development. The new version has some issue for compiling.

There are very few modern, open source P2P libraries available, particularly in Go. Overall, go-libp2p is quite good and it’s well suited to our objectives. The best way to get your code environment set up is to clone this entire library and write your code within it. We used

```
$ go mod tidy

$ go mod vendor
```

to make sure we have entire library on local. At the time of developing this project I had to do a few fixes on libp2p2 to be able to use it in my codes. I pushed changes in branch p2p this repo. So, if you can't build the libp2p, can clone the branch libp2p here and copy it on vendors folder. (overwrite with what is there)


## install logger
BDC uses viper as logging service.

`
env GO111MODULE=on go get github.com/spf13/viper
`
# Test
to test all packages use command below

```
$ go test ./src/...
```

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

```
$ go test ./src/... -coverprofile=./c.out && go tool cover -html=c.out && unlink ./c.out
```

# P2P Networking
Starts a libp2p node. Sets up pubsub, subscribes to "blocks" and "transactions" topics. Connecting to another node is totally automate. So, any time you add a new node to network, as long as other nodes can see the IP address, they will detect new nodes and connect to each other. The new node will sync itself with latest network state.
Some of the key features of libp2p are as follows:
- encrypted connections
- able to work offline
- It's easy to switch between transporters 
- very flexible for switching storages (TCP, UDP, Relay, ...)
- can be used as private and public network
- supports different routing (kademlia, DHT, ...) and some other useful Mock routing

read more here https://github.com/libp2p/go-libp2p
# Block Storage
BDC uses leveldb as block storage. This storage are handled by go-ipfs-blockservice. But for indexing the blocks, we use another db.

# Mining
Mining is a proof-of-work algorithm that hashes a random nonce using sha256, seeking a target solution. To enable the mining for node, set Mining Enabled to true in configurations.

# Block structure
The BDC block structure is like this:

```golang
type BlockHeader struct {
	Version    string
	PrevHash   hash.Hash
	MerkleRoot hash.Hash
	Timestamp  int64
	Nonce      int64
	Miner      string
	Difficulty uint32
	Memo       string
}

type Block struct {
	Height       uint64
	Hash         hash.Hash
	PrevCid      *cid.Cid
	Header       BlockHeader
	Reward       *big.Float
	TxsCount     uint64
	Transactions []*transaction.Transaction
}
```

As you may notice it is very similar to Bitcoin. We allow some optional data as a message for each block.

Each block received over network is processed, and saved if it is valid. Based on longest chain, we reload blocks.

# Transaction
The BDC's transaction structure is very similar to Ethereum blockchain. 

```golang
type Transaction struct {
	ID        hash.Hash
	Nonce     uint64
	PublicKey []byte
	Signature []byte
	Timestamp int64
	From      string
	To        string
	Fee       uint64
	Value     float64
	Data      string
}

```

** To ensure protecting against double spending and replay attack, we use Nonce for each transaction which is same idea as ethereum

# Wallet
The CLI is able to create new wallet and send transaction. BDC supports wallet set which can manage a set of wallets and also add new wallet to the list.

# CLI

CLI starts up an http server, provides command line RPC interface. 
Cli must be built and run separately.

```
$ go build -o bdc-cli ./cli
```

and you can use it like:

```
$ ./bdc-cli help
```

```
NAME:
   bdc-cli - rpc client for badcoin

USAGE:
   bdc-cli [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   status, stat      shows connection status
   sendtx, tx        send a transaction
   newaddress, addr  get new address
   info, i           shows blockchain information
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

# Server Endpoints

 url  			  |  method   | 	parameters 	               | 	description	                      |
 -----------------|-----------|--------------------------------|--------------------------------------|
 /Info            | Get       | -                              |return BDC node info                  |
 /Block           | Get       | height                         |returns a certain block heigh details |
 /Genesis         | Get       | -                              |returns genesis block                 |
 /Tx/Send         | Post      | to,value,data                  |send a new transaction (miner wallet) |
 /Tx/Signed/Send  | Post      | to,value,pubkey,signature,data |send a new signed transaction         |
 /Address/New     | Post      | -                              |generate a new address                |