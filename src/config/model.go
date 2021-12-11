package config

const (
	//storage types
	STORAGE_LEVELDB = 1
)

// Configurations exported
type Configurations struct {
	Env        string
	Name       string
	ConfigFile string
	ID         string
	Genesis    Genesis
	MiningSet  MiningSet
	RpcSet     RpcSet
	Storage    Storage
}

//Genesis for genesis block options
type Genesis struct {
	Height uint64
	Nonce  int64
	Reward uint64
}

// MiningSet mining config
type MiningSet struct {
	Enabled                     bool
	ExpectedMiningTimeInSeconds uint64
}

// RpcSet rpc server config
type RpcSet struct {
	Enabled bool
	Port    string
}

// data storage config
type Collections struct {
	Blocks     string
	BlockIndex string
	UTXO       string
	TXsMemPool string
	Stats      string
}
type Storage struct {
	Type        uint8
	DBName      string
	Collections Collections
}
