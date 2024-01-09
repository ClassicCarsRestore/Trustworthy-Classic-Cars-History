"use strict";

import fs from "fs";
import { fileURLToPath } from "url";
import { dirname, path } from "path";
import yaml from "js-yaml";
import nodemailer from 'nodemailer';
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
import dotenv from "dotenv";
dotenv.config();

const transporter = nodemailer.createTransport({
  service: 'Gmail',
  auth: {
    user: process.env.EMAIL_EMAIL,
    pass: process.env.EMAIL_PASS,
  },
});

export function buildCCP() {
  // load the common connection configuration file
  const ccpPath = path.resolve(__dirname, "..", "network.yaml");
  const fileExists = fs.existsSync(ccpPath);
  if (!fileExists) {
    throw new Error(`no such file or directory: ${ccpPath}`);
  }
  const ccp = yaml.load(fs.readFileSync(ccpPath, "utf8"));
  console.log(`Loaded the network configuration located at ${ccpPath}`);
  console.log(ccp);
  return ccp;
}

export async function buildWallet(Wallets, walletPath) {
  // Create a new  wallet : Note that wallet is for managing identities.
  let wallet;
  if (walletPath) {
    wallet = await Wallets.newFileSystemWallet(walletPath);
    console.log(`Built a file system wallet at ${walletPath}`);
  } else {
    wallet = await Wallets.newInMemoryWallet();
    console.log("Built an in memory wallet");
  }
  return wallet;
}

export function prettyJSONString(inputString) {
  if (inputString) {
    return JSON.stringify(JSON.parse(inputString), null, 2);
  } else {
    return inputString;
  }
}

export async function sendEmail(recipient, subject, message) {
  console.log(process.env.EMAIL_EMAIL);
  const mailOptions = {
    from: 'raimundo.branco@gmail.com',
    to: recipient,
    subject: subject,
    text: message,
  };

  transporter.sendMail(mailOptions, (error, info) => {
    if (error) {
      console.error('Error sending email: ', error);
    } else {
      console.log('Email sent: ' + info.response);
    }
  });

}

