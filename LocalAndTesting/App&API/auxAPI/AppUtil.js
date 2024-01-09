"use strict";

import fs from "fs";
import { fileURLToPath } from 'url';
import { dirname}  from 'path';
import path from "path";
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export function buildCCPOrg1() {
  // load the common connection configuration file
  const ccpPath = path.resolve(
    __dirname,
    "..",
    "..",
    "TestNetwork",
    "organizations",
    "peerOrganizations",
    "org1.example.com",
    "connection-org1.json"
  );
  //const ccpPath = path.resolve(__dirname, "..","network.yaml");

  const fileExists = fs.existsSync(ccpPath);
  if (!fileExists) {
    throw new Error(`no such file or directory: ${ccpPath}`);
  }

  const contents = fs.readFileSync(ccpPath, "utf8");

  // build a JSON object from the file contents
  const ccp = JSON.parse(contents);
  //const ccp = yaml.load(fs.readFileSync(ccpPath, "utf8"));

  console.log(`Loaded the network configuration located at ${ccpPath}`);
  //ADICIONADA POR MIM
  console.log(ccp)
  return ccp;
};

export function buildCCPOrg2() {
  // load the common connection configuration file
  const ccpPath = path.resolve(
    __dirname,
    "..",
    "..",
    "TestNetwork",
    "organizations",
    "peerOrganizations",
    "org2.example.com",
    "connection-org2.json"
  );
  //const ccpPath = path.resolve(__dirname, "..", "network.yaml");
  const fileExists = fs.existsSync(ccpPath);
  if (!fileExists) {
    throw new Error(`no such file or directory: ${ccpPath}`);
  }

  const contents = fs.readFileSync(ccpPath, "utf8");

  // build a JSON object from the file contents
  const ccp = JSON.parse(contents);
  //const ccp = yaml.load(fs.readFileSync(ccpPath, "utf8"));

  console.log(`Loaded the network configuration located at ${ccpPath}`);
  return ccp;
};

export function buildCCPOrg3() {
  // load the common connection configuration file
  const ccpPath = path.resolve(
    __dirname,
    "..",
    "..",
    "TestNetwork",
    "organizations",
    "peerOrganizations",
    "org3.example.com",
    "connection-org3.json"
  );
  //const ccpPath = path.resolve(__dirname, "..", "network.yaml");
  const fileExists = fs.existsSync(ccpPath);
  if (!fileExists) {
    throw new Error(`no such file or directory: ${ccpPath}`);
  }

  const contents = fs.readFileSync(ccpPath, "utf8");

  // build a JSON object from the file contents
  const ccp = JSON.parse(contents);
  //const ccp = yaml.load(fs.readFileSync(ccpPath, "utf8"));

  console.log(`Loaded the network configuration located at ${ccpPath}`);
  return ccp;
};

export async function buildWallet (Wallets, walletPath) {
	// Create a new  wallet : Note that wallet is for managing identities.
	let wallet;
	if (walletPath) {
		wallet = await Wallets.newFileSystemWallet(walletPath);
		console.log(`Built a file system wallet at ${walletPath}`);
	} else {
		wallet = await Wallets.newInMemoryWallet();
		console.log('Built an in memory wallet');
	}

	return wallet;
};

export function prettyJSONString(inputString) {
	if (inputString) {
		 return JSON.stringify(JSON.parse(inputString), null, 2);
	}
	else {
		 return inputString;
	}
}
