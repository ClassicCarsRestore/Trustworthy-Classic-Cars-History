#!/bin/bash
set -ex
# Bring the test network down
pushd ../TestNetwork
./network.sh down
popd

# clean out any old identites in the wallets
rm -rf javascript/wallet/*
rm -rf java/wallet/*
rm -rf typescript/wallet/*
