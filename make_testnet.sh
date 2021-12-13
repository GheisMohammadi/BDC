go build
rm -rf ./testnet/node1/badcoin
rm -rf ./testnet/node2/badcoin
cp ./badcoin ./testnet/node1/badcoin
cp ./badcoin ./testnet/node2/badcoin