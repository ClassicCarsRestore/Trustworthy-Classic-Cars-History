#!/bin/bash
# Exit on first error
#e da exit quando da erro, x da log a todos os comandos a serem executados
#(com para debug)
set -ex

cd .

#Assim ${1} le o primeiro arugmento passado quando corre o script
printf "Chaincode to deploy: ${1}\n"

#Assim e para ler mesmo do terminal
#read cc_name 


# clean out any old identites in the wallets
rm -rf javascript/wallet/*
rm -rf java/wallet/*
rm -rf typescript/wallet/*
rm -rf go/wallet/*

starttime=$(date +%s)

# launch network; create channel and join peer to channel
pushd ../TestNetwork
./network.sh down
./network.sh up createChannel -ca 
#Using couch db
#access: http://localhost:5984/_utils
#./network.sh up createChannel -ca -s couchdb
./network.sh deployCC -ccn ${1} -ccp ../Chaincode/chaincode-go/ -ccl go
#popd

cat <<EOF

Total setup execution time : $(($(date +%s) - starttime)) secs ...



Congrats!!!! ${1} Was succesfully deployed!

EOF