#!/bin/bash

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/../TestNetwork/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/../TestNetwork/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
#export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/../TestNetwork/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
export CORE_PEER_ADDRESS=localhost:9051
echo "Did you use the . before ./set-env.sh? If yes then we are good :)"
