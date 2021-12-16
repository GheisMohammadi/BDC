echo "building..."
go build
echo "deleting current nodes data..."
rm -rf ./testnet/node1/badcoin
rm -rf ./testnet/node1/data
rm -rf ./testnet/node2/badcoin
rm -rf ./testnet/node2/data
echo "installing new nodes..."
cp ./badcoin ./testnet/node1/badcoin
cp ./badcoin ./testnet/node2/badcoin
echo "done!"