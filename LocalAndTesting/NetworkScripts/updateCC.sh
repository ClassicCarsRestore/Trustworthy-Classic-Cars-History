#!/bin/bash
# Exit on first error
#e da exit quando da erro, x da log a todos os comandos a serem executados
#(com para debug)
set -ex

cd .

#Assim ${1} le o primeiro arugmento passado quando corre o script
printf "Chaincode to update: ${1}, version ${2}, sequence ${3}\n"

starttime=$(date +%s)

pushd ../TestNetwork
./network.sh deployCC -ccn ${1} -ccp ../Chaincode/chaincode-go/ -ccl go -ccv ${2} -ccs ${3}
#popd

cat <<EOF

Total setup execution time : $(($(date +%s) - starttime)) secs ...



Congrats!!!! ${1} Was succesfully UPDATED!

EOF