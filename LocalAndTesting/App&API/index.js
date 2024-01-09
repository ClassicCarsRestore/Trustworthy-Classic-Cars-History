import express from "express";
import bodyParser from "body-parser";
// Setting for Hyperledger Fabric
import { Gateway, Wallets } from "fabric-network";
import FabricCAServices from "fabric-ca-client";
import fs from "fs";
//JWT
import jwt from "jsonwebtoken";
import bearerToken from "express-bearer-token";

//MongoDB
import { MongoClient } from "mongodb";
import dotenv from "dotenv";
dotenv.config();
//Bcrypt
import bcrypt from "bcrypt";

//Fabric local modules
import {
  buildCAClient,
  registerAndEnrollUser,
  enrollAdmin,
} from "./auxAPI/CAUtil.js";
import { buildCCPOrg1, buildCCPOrg2, buildCCPOrg3, buildWallet } from "./auxAPI/AppUtil.js";

import { fileURLToPath } from "url";
import { dirname } from "path";
import path from "path";

import { create } from "ipfs-http-client";
//import fs from 'fs'
import fileUpload from "express-fileupload";

//Swagger
import swaggerUi from "swagger-ui-express";
import yaml from "js-yaml";

import { BlockDecoder } from 'fabric-common';
import { X509Certificate } from 'crypto';

//Fabric Global Variables
//let contract;
//let gateway;
let wallet;
let wallet2;
let wallet3;
let ccp;
let ccp2;
let ccp3;

//TESTE
let gateways = [];

//MONGODB SETTINGS
//TALVEZ USAR REDIS PARA USER DATABASE
//https://medium.com/swlh/session-management-in-nodejs-using-redis-as-session-store-64186112aa9
let usersdb;

//IPFS
let ipfs;
const ipfsUrl = "https://ipfs.io/ipfs/";

const app = express();
app.use(bodyParser.json());
app.use(fileUpload());

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// set secret variable
app.set("secret", process.env.JWT_SECRET);

app.use(bearerToken());

// Read the Swagger YAML file
const swaggerDocument = yaml.load(fs.readFileSync("./swagger.yaml", "utf8"));

// Validate routes against Swagger documentation
app.use("/api-docs", swaggerUi.serve, swaggerUi.setup(swaggerDocument));

// Import the NFTStorage class and File constructor from the 'nft.storage' package
import { NFTStorage, File  } from 'nft.storage';
const NFT_STORAGE_KEY = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkaWQ6ZXRocjoweDYzREY5MUVkRGQ4Nzg0NTQ0Y2ExRjQ5ZjQxNDE4OTYxOWY5MTQwQTUiLCJpc3MiOiJuZnQtc3RvcmFnZSIsImlhdCI6MTY5Mzc2OTIyODA3NiwibmFtZSI6InRlc3QifQ.PBwOi9p8QcKDi5UInhSZ20HbBNGS4XCi1w4oMqE93HE';
let nftstorage;

app.use(async (req, res, next) => {
  if (
    req.originalUrl.indexOf("api/Users") >= 0 ||
    req.originalUrl.indexOf("api/Users/Login") >= 0 ||
    req.originalUrl.indexOf("/docs") >= 0
  ) {
    return next();
  }
  const token = req.token;
  jwt.verify(token, app.get("secret"), async (err, decoded) => {
    if (err) {
      console.log(`Error ================:${err}`);
      res.status(400).send({
        success: false,
        message:
          "Failed to authenticate token. Make sure to include the " +
          "token returned from /users call in the authorization header " +
          " as a Bearer token",
      });
      return;
    } else {
      req.email = decoded.email;
      req.orgName = decoded.orgName;



      console.log("===USER INFO===");
      console.log(`EMAIL: ${req.email}`);
      console.log(`ORGANIZATION: ${req.orgName}`);
      console.log("===============");
      return next();
    }
  });
});

app.listen(8393, async () => {
  try {
    //IPFS node init
    ipfs = await create();
    nftstorage = new NFTStorage({ token: NFT_STORAGE_KEY })

    //Connect to MongoDB
    connectToMongo();

    // build an in memory object with the network configuration (also known as a connection profile)
    //Org1 for normal users/clients/classic owners
    ccp = buildCCPOrg1();
    //Org2 for workshops/restoration shops/garages
    ccp2 = buildCCPOrg2();
    //Org3 for certification authorities and other similar entities
    ccp3 = buildCCPOrg3();

    // build an instance of the fabric ca services client based on
    // the information in the network configuration
    const caClient = buildCAClient(
      FabricCAServices,
      ccp,
      "ca.org1.example.com"
    );

    const caClient2 = buildCAClient(
      FabricCAServices,
      ccp2,
      "ca.org2.example.com"
    );

    const caClient3 = buildCAClient(
      FabricCAServices,
      ccp3,
      "ca.org3.example.com"
    );

    // setup the wallet to hold the credentials of the application user
    wallet = await buildWallet(Wallets, path.join(__dirname, "org1-wallet"));
    wallet2 = await buildWallet(Wallets, path.join(__dirname, "org2-wallet"));
    wallet3 = await buildWallet(Wallets, path.join(__dirname, "org3-wallet"));

    // in a real application this would be done on an administrative flow, and only once
    await enrollAdmin(caClient, wallet, "Org1MSP");
    await enrollAdmin(caClient2, wallet2, "Org2MSP");
    await enrollAdmin(caClient3, wallet3, "Org3MSP");

    // in a real application this would be done only when a new user was required to be added
    // and would be part of an administrative flow
    await registerAndEnrollUser(
      caClient,
      wallet,
      "Org1MSP",
      "appUser",
      "org1.department1"
    );

    // Create a new gateway instance for interacting with the fabric network.
    // In a real application this would be done as the backend server session is setup for
    // a user that has been verified.
    const gateway = new Gateway();

    try {
      // setup the gateway instance
      // The user will now be able to create connections to the fabric network and be able to
      // submit transactions and query. All transactions submitted by this gateway will be
      // signed by this user using the credentials stored in the wallet.
      await gateway.connect(ccp, {
        wallet,
        identity: "appUser",
        discovery: { enabled: true, asLocalhost: true }, // using asLocalhost as this gateway is using a fabric network deployed locally
      });

      // Build a network instance based on the channel where the smart contract is deployed
      const network = await gateway.getNetwork("mychannel");

      // Get the contract from the network.
      const contract = network.getContract("classicCars");

      console.log(
        "\n--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger"
      );
      await contract.submitTransaction("InitLedger");
      console.log("*** Result: committed");
    } finally {
      // Disconnect from the gateway when the application is closing
      // This will close all connections to the network
      // gateway.disconnect();
    }
  } catch (error) {
    console.error(`******** FAILED to run the application: ${error}`);
  }
});

/**
 * @POST
 * @async
 * @description Registers a user with a given email, orgName and email present in the body request
 */
app.post("/api/Users", async (req, res) => {
  if (!validNArguments(req, 5)) {
    return errorArguments(res);
  }
  let email = req.body.email;
  let password = req.body.password;
  //TALVEZ REPEAT PASSWORD
  let orgName = req.body.orgname;
  let firstname = req.body.firstname;
  let surname = req.body.surname;
  if (!email) {
    return res.status(400).json(emptyFieldError("'email'"));
  }
  if (!password) {
    return res.status(400).json(emptyFieldError("'password'"));
  }
  if (!orgName) {
    return res.status(400).json(emptyFieldError("'orgName'"));
  }
  if (!firstname) {
    return res.status(400).json(emptyFieldError("'firstname'"));
  }
  if (!surname) {
    return res.status(400).json(emptyFieldError("'surname'"));
  }
  let userInWallet = await isUserRegistered(email, orgName);
  if (userInWallet) {
    return confictErrorUser(email, res);
  }
  try {
    let ccpUser;
    let orgMspIdUser;
    let caHostNameUser;
    let userWallet;
    let affiliationUser;
    let role;
    if (orgName == "Org1") {
      ccpUser = ccp;
      orgMspIdUser = "Org1MSP";
      caHostNameUser = "ca.org1.example.com";
      userWallet = wallet;
      affiliationUser = "org1.department1";
      role = "client"
    } else if (orgName == "Org2") {
      ccpUser = ccp2;
      orgMspIdUser = "Org2MSP";
      caHostNameUser = "ca.org2.example.com";
      userWallet = wallet2;
      affiliationUser = "org2.department1";
      role = "workshop"
    }
    else if (orgName == "Org3") {
      ccpUser = ccp3;
      orgMspIdUser = "Org3MSP";
      caHostNameUser = "ca.org3.example.com";
      userWallet = wallet3;
      affiliationUser = "org3.department1";
      role = "certifier"
    }

    const caClientUser = buildCAClient(
      FabricCAServices,
      ccpUser,
      caHostNameUser
    );
    const walletUser = userWallet;

    await registerAndEnrollUser(
      caClientUser,
      walletUser,
      orgMspIdUser,
      email,
      affiliationUser,
      "client"
    );

    //MongoDB
    //10 is the salt
    const hashedpw = await bcrypt.hash(password, 10);
    usersdb.insertOne({
      email: email,
      password: hashedpw,
      org: orgName,
      role: role,
      firstname: firstname,
      surname: surname
    });

    let result = {
      email: email,
      success: true,
    };

    res.status(200).json(result);
  } catch (error) {
    internalServerError("evaluate", error, res);
  }
});

/**
 * @POST
 * @async
 * @description Logins a user with a given email, orgName and email present in the body request
 * @returns a message containing a JWT token
 */
app.post("/api/Users/Login", async (req, res) => {
  if (!validNArguments(req, 3)) {
    return errorArguments(res);
  }
  let email = req.body.email;
  let password = req.body.password;
  let orgName = req.body.orgname;
  if (!email) {
    return res.status(400).json(emptyFieldError("'email'"));
  }
  if (!password) {
    return res.status(400).json(emptyFieldError("'password'"));
  }
  if (!orgName) {
    return res.status(400).json(emptyFieldError("'orgName'"));
  }

  //MongoDB
  const dbUser = await usersdb.findOne({ email: email });
  if (!dbUser) {
    return notFoundErrorUser(email, orgName, res);
  }
  const userAllowed = await bcrypt.compare(password, dbUser.password);
  let userInWallet = await isUserRegistered(dbUser.email, orgName);

  if (userInWallet && userAllowed) {
    let token = jwt.sign(
      {
        exp: Math.floor(Date.now() / 1000) + 3600,
        email: dbUser.email,
        orgName: dbUser.org,
      },
      app.get("secret")
    );

    res.json({
      success: true,
      message: { email: dbUser.email, token: token },
    });
  } else if (!userAllowed) {
    res.status(401).json({
      success: false,
      message: `The password is incorrect. Try again.`,
    });
  } else {
    notFoundErrorUser(email, orgName, res);
  }
});

/**
 * @POST
 * @async
 * @description Auxiliar function to test file upload to IPFS
 * @returns a sucess or fail message
 */
app.post("/api/addImage", async (req, res) => {
  try {
    // create a new NFTStorage client using our API key
    let photosIds = [];
    if (!req.files) {
      return res.status(400).json({ error: "No file uploaded" });
    } else {
      console.log("Uploading files...");
      for (const file of req.files.file) {
        console.log(file);
        const imageFile = new File([file.data], 'document', { type: 'image/*' });
        const metadata = await nftstorage.store({
          name: 'Random Name',
          description: 'Random Description',
          image: imageFile
        });
        let cid = metadata.embed().image.href.toString();
        console.log(cid);
        photosIds.push(cid);
      }
      console.log(photosIds);
      console.log("Files were uploaded!");
      return res.status(200).json({ Sucess: "File uploaded" });
    }
  } catch (error) {
    console.log(error);
  }
});

/**
 * @GET
 * @async
 * @description Auxiliar function to get all the users registered in the system
 * @returns a sucess or fail message
 */
app.get("/api/Users/GetAll", async (req, res) => {
  try {
    const projection = { projection: { _id: 0, firstname: 1, surname:1, email: 1, role: 1 } };
    const users = await usersdb.find({}, projection).toArray();
    //If desired as a simple array
    //const emails = users.map((user) => user.email);
    console.log(users);
    return res.status(200).json(users);
  } catch (error) {
    console.log(error);
  }
});

/**
 * @GET
 * @async
 * @description Auxiliar function to get all the user info
 * @returns a sucess or fail message
 */
app.get("/api/Users/Get/:email", async (req, res) => {
  try {
    const projection = { projection: { _id: 0, firstname: 1, surname:1, email: 1, role: 1 } };
    const dbUser = await usersdb.findOne({ email: req.params.email }, projection);
    if (!dbUser) {
      console.log(`User with email ${req.params.email} does not exist in the system`);
      res.status(404).json({
        success: false,
        message: `User with email ${req.params.email} is not registered in the system.`,
      });
    }
    console.log(dbUser);
    return res.status(200).json(dbUser);
  } catch (error) {
    console.log(error);
  }
});

/**
 * @GET
 * @async
 * @description Gets all the available classics in the system (works as homepage for admins)
 */
app.get("/api/Classics/GetAll", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.evaluateTransaction("GetAllClassics");
    console.log(
      `Transaction has been evaluated, result is: ${result.toString()}`
    );
    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    internalServerError("evaluate", error, res);
  }
});

/**
 * @POST
 * @async
 * @description Creates a new classic with details got from the request body, and adds this classic
 * to the ledger
 */
app.post("/api/Classics/Create", async (req, res) => {
  try {
    if (!validNArguments(req, 8)) {
      return errorArguments(res);
    }
    if (!req.body.chassisNo) {
      return res.status(400).json(emptyFieldError("'VIN/Chassis number'"));
    }
    if (!req.body.ownerEmail) {
      return res.status(400).json(emptyFieldError("'Owner email'"));
    }

    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const currentDate = new Date();
    await contract.submitTransaction(
      "CreateClassic",
      req.body.make,
      req.body.model,
      req.body.year,
      req.body.licencePlate,
      req.body.country,
      req.body.chassisNo,
      req.body.engineNo,
      req.body.ownerEmail,
      currentDate.toISOString()
    );
    console.log(
      `The car with chassis number ${req.body.chassisNo} was created.`
    );

    res.status(200).json({
      success: true,
      message:
        "Classic with chassis number " +
        req.body.chassisNo +
        " was added successfully!",
    });
  } catch (error) {
    if (
      error.responses &&
      error.responses.length > 0 &&
      error.responses[0].hasOwnProperty("response")
    ) {
      if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      }
      else if (error.responses[0].response.message === "409") {
        confictErrorClassic(req.body.chassisNo, error, res);
      }
    } else {
      internalServerError("submit", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description Changes the email of the classic's owner (works as ownership change);
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.put("/api/Classics/UpdateEmail/:chassisNo/:newEmail", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    await contract.submitTransaction(
      "UpdateClassicEmail",
      req.params.chassisNo,
      req.params.newEmail
    );
    res.status(200).json({
      success: true,
      message: `The car with chassis number ${req.params.chassisNo} was updated to the email ${req.params.newEmail}.`,
    });
    console.log(
      `The car with chassis number ${req.params.chassisNo} was updated to the email ${req.params.newEmail}.`
    );
  } catch (error) {
    if (error.responses[0].response.message === "404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error.responses[0].response.message === "403") {
      forbiddenError(res);
    } else {
      internalServerError("submit", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description Updates the classic's documentation by adding a new restoration procedure with details
 * got from the request body, including any type of file
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.post("/api/Restorations/Create/:chassisNo", async (req, res) => {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classicCars");
  try {
    access = await contract.evaluateTransaction(
      "HasModifierAccess",
      req.params.chassisNo
    );
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
  if (access == "true") {
    try {
      if (!validNArguments(req, 2)) {
        return errorArguments(res);
      }
      if (!req.body.title) {
        return res
          .status(400)
          .json(emptyFieldError("'Title of the restoration'"));
      }
      //===>IPFS
      let photosIds = [];
      let stringified;
      if (!req.files) {
        console.log("No files were uploaded");
      } else if (!Array.isArray(req.files.file)) {
        console.log("Uploading a file...");
        const file = req.files.file;
        console.log(file);
        const imageFile = new File([file.data], req.params.chassisNo, { type: 'image/*' });
        const metadata = await nftstorage.store({
          name: req.body.title,
          description: req.body.description,
          image: imageFile
        });
        let cid = metadata.embed().image.href.toString();
        console.log(cid);
        photosIds.push(cid);
        console.log("File was uploaded!");
      } else {
        console.log("Uploading files...");
        for (const file of req.files.file) {
          console.log(file);
          const imageFile = new File([file.data], req.params.chassisNo, { type: 'image/*' });
          const metadata = await nftstorage.store({
            name: req.body.title,
            description: req.body.description,
            image: imageFile
          });
          let cid = metadata.embed().image.href.toString();
          console.log(cid);
          photosIds.push(cid);
        }
        console.log("Files were uploaded!");
      }
      stringified = JSON.stringify(photosIds);
      const currentDate = new Date();
      await contract.submitTransaction(
        "CreateRestorationStep",
        req.params.chassisNo,
        req.body.title,
        req.body.description,
        stringified,
        currentDate.toISOString()
      );
      res.status(200).json({
        success: true,
        message: `The car with chassis number ${req.params.chassisNo} was updated with a new restoration procedure.`,
      });
      console.log(
        `The car with chassis number ${req.params.chassisNo} was updated with a new restoration procedure.`
      );
    } catch (error) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else {
        internalServerError("submit", error, res);
      }
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets a specific restoration procedure of a classic
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.get("/api/Restorations/Get/:chassisNo/:stepId", async (req, res) => {
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classicCars");
  try {
    const result = await contract.evaluateTransaction(
      "GetStep",
      req.params.chassisNo,
      req.params.stepId
    );
    console.log(
      `Transaction has been evaluated, result is: ${result.toString()}`
    );
    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else if (error == "Error: 404nostep") {
      res.status(404).json({
        success: false,
        message: `The step ${req.params.stepId} was not found.`,
      });
      console.log(`The step ${req.params.stepId} was not found.`);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description Updates a specific restoration procedure of a classic
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.put("/api/Restorations/Update/:chassisNo", async (req, res) => {
  try {
    if (!validNArguments(req, 3)) {
      return errorArguments(res);
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    await contract.submitTransaction(
      "UpdateStep",
      req.params.chassisNo,
      req.body.stepId,
      req.body.newTitle,
      req.body.newDescription
    );
    res.status(200).json({
      success: true,
      message: `The step ${req.body.stepId} was updated sucessfully.`,
    });
    console.log(`The restoration ${req.body.stepId} was updated.`);
  } catch (error) {
    if (error.responses[0].response.message === "404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error.responses[0].response.message === "403") {
      forbiddenError(res);
    } else if (error.responses[0].response.message === "404madeby") {
      res.status(404).json({
        success: false,
        message: `Not authorized to update step ${req.body.stepId} since you are who made it originally.`,
      });
      console.log(
        `The restoration ${req.body.stepId} was NOT updated. Not who originally made it.`
      );
    } else if (error.responses[0].response.message === "404nostep") {
      res.status(404).json({
        success: false,
        message: `The step ${req.body.stepId} was not found.`,
      });
      console.log(`The step ${req.body.stepId} was not found.`);
    } else {
      internalServerError("submit", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description Updates a specific restoration procedure of a classic with new photos/documents
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.put("/api/Restorations/Update/:chassisNo/Photos", async (req, res) => {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classicCars");
  try {
    access = await contract.evaluateTransaction(
      "HasModifierAccess",
      req.params.chassisNo
    );
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
  if (access == "true") {
    try {
      if (!validNArguments(req, 1)) {
        return errorArguments(res);
      }
      let photosIds = [];
      let stringified;
      if (!req.files) {
        console.log("No files were uploaded");
        return res.status(400).json({
          success: false,
          message: "Please add at least one file!",
        });
      } else if (!Array.isArray(req.files.file)) {
        console.log("Uploading a file...");
        const file = req.files.file;
        console.log(file);
        const imageFile = new File([file.data], req.params.chassisNo, { type: 'image/*' });
        const metadata = await nftstorage.store({
          name: req.params.chassisNo,
          description: req.body.stepId,
          image: imageFile
        });
        let cid = metadata.embed().image.href.toString();
        console.log(cid);
        photosIds.push(cid);
        console.log("File was uploaded!");
      } else {
        console.log("Uploading files...");
        for (const file of req.files.file) {
          console.log(file);
          const imageFile = new File([file.data], req.params.chassisNo, { type: 'image/*' });
          const metadata = await nftstorage.store({
            name: req.params.chassisNo,
            description: req.body.stepId,
            image: imageFile
          });
          let cid = metadata.embed().image.href.toString();
          console.log(cid);
          photosIds.push(cid);
        }
        console.log("Files were uploaded!");
      }
      stringified = JSON.stringify(photosIds);
      await contract.submitTransaction(
        "UpdateStepPhotos",
        req.params.chassisNo,
        req.body.stepId,
        stringified
      );
      res.status(200).json({
        success: true,
        message: `The photos/documents of step ${req.body.stepId} were updated sucessfully.`,
      });
      console.log(`The restoration ${req.body.stepId} was updated.`);
    } catch (error) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else if (error.responses[0].response.message === "404madeby") {
        res.status(404).json({
          success: false,
          message: `Not authorized to update step ${req.body.stepId} since you are who made it originally.`,
        });
        console.log(
          `The restoration ${req.body.stepId} was NOT updated. Not who originally made it.`
        );
      } else if (error.responses[0].response.message === "404nostep") {
        res.status(404).json({
          success: false,
          message: `The step ${req.body.stepId} was not found.`,
        });
        console.log(`The step ${req.body.stepId} was not found.`);
      } else {
        internalServerError("submit", error, res);
      }
    }
  }
});

/**
 * @PUT
 * @async
 * @description Updates a specific restoration procedure of a classic with new title, description and photos/documents
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.put("/api/Restorations/UpdateAndPhotos/:chassisNo", async (req, res) => {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classicCars");
  try {
    access = await contract.evaluateTransaction(
      "HasModifierAccess",
      req.params.chassisNo
    );
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
  if (access == "true") {
    try {
      if (!validNArguments(req, 3)) {
        return errorArguments(res);
      }
      let photosIds = [];
      let stringified;
      if (!req.files) {
        console.log("No files were uploaded");
      } else if (!Array.isArray(req.files.file)) {
        console.log("Uploading a file...");
        const file = req.files.file;
        console.log(file);
        const imageFile = new File([file.data], req.params.chassisNo, { type: 'image/*' });
        const metadata = await nftstorage.store({
          name: req.params.chassisNo,
          description: req.body.stepId,
          image: imageFile
        });
        const cid = metadata.embed().image.href.toString();
        console.log(cid);
        photosIds.push(cid);
        console.log("File was uploaded!");
      } else {
        console.log("Uploading files...");
        for (const file of req.files.file) {
          console.log(file);
          const imageFile = new File([file.data], req.params.chassisNo, { type: 'image/*' });
          const metadata = await nftstorage.store({
            name: req.params.chassisNo,
            description: req.body.stepId,
            image: imageFile
          });
          const cid = metadata.embed().image.href.toString();
          console.log(cid);
          photosIds.push(cid);
        }
        console.log("Files were uploaded!");
      }
      stringified = JSON.stringify(photosIds);
      await contract.submitTransaction(
        "UpdateStepAndPhotos",
        req.params.chassisNo,
        req.body.stepId,
        req.body.newTitle,
        req.body.newDescription,
        stringified
      );
      res.status(200).json({
        success: true,
        message: `The photos/documents of step ${req.body.stepId} were updated sucessfully.`,
      });
      console.log(`The restoration ${req.body.stepId} was updated.`);
    } catch (error) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else if (error.responses[0].response.message === "404madeby") {
        res.status(404).json({
          success: false,
          message: `Not authorized to update step ${req.body.stepId} since you are who made it originally.`,
        });
        console.log(
          `The restoration ${req.body.stepId} was NOT updated. Not who originally made it.`
        );
      } else if (error.responses[0].response.message === "404nostep") {
        res.status(404).json({
          success: false,
          message: `The step ${req.body.stepId} was not found.`,
        });
        console.log(`The step ${req.body.stepId} was not found.`);
      } else {
        internalServerError("submit", error, res);
      }
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets the details of a specified classic with the chassisNo got form the url parameters
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.get("/api/Classics/Get/:chassisNo", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const currentDate = new Date();
    const result = await contract.evaluateTransaction(
      "ReadClassicAsViewer",
      req.params.chassisNo,
      currentDate.toISOString()
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);
    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description Updates a specific classic with new documentation
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.put("/api/Classics/AddDocument/:chassisNo", async (req, res) => {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classicCars");
  try {
    //MAYBE CHANGE THE NECESSARY ACCESS
    access = await contract.evaluateTransaction(
      "HasDocumenterAccess",
      req.params.chassisNo
    );
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
  if (access == "true") {
    try {
      if (!req.body.name) {
        return res.status(400).json(emptyFieldError("'Name of the document'"));
      }
      if (!req.files) {
        return res.status(400).json(emptyFieldError("'No file was provided.'"));
      }

      console.log("Uploading documents...");
      let doc = req.files.file;
      console.log(doc);
      const imageFile = new File([doc.data], req.body.name, { type: 'image/*' });
      const metadata = await nftstorage.store({
        name: req.body.name,
        description: req.body.name,
        image: imageFile
      });
      let docId = metadata.embed().image.href.toString();
      console.log(docId);
      console.log("Document was uploaded!");
      await contract.submitTransaction(
        "AddDocument",
        req.params.chassisNo,
        req.body.name,
        docId
      );
      res.status(200).json({
        success: true,
        message: `A new document was added to the car with chassis number ${req.params.chassisNo}.`,
      });
      console.log(
        `A new document was added to the car with chassis number ${req.params.chassisNo}.`
      );
    } catch (error) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else {
        internalServerError("submit", error, res);
      }
      console.error(error);
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets the history of a specified classic with the chassisNo got form the url parameters
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.get("/api/Classics/History/:chassisNo", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const currentDate = new Date();
    const result = await contract.evaluateTransaction(
      "GetClassicHistory2",
      req.params.chassisNo,
      currentDate.toISOString()
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);

    const rr = JSON.parse(result);
    res.status(200).send(rr);

    //em json mas com backslashes
    //res.status(200).json({ response: result.toString() });
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets all the classics owned by a specific user
 * @returns List of classics if the user making the request ownes at least one classic, error otherwise
 */
app.get("/api/Classics/GetByOwner/", async (req, res) => {
  try {
    const email = req.email;
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.evaluateTransaction(
      "QueryClassicsByOwner",
      email
    );
    console.log(
      `Transaction has been evaluated, result is: ${result.toString()}`
    );
    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    if (error == "Error: 404") {
      let response = {
        success: false,
        message: `404 - User ${req.email} has no classics registered in the system`,
      };
      res.status(404).json(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

app.get("/api/Classics/GetByModifier/", async (req, res) => {
  try {
    const email = req.email;
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.evaluateTransaction(
      "QueryClassicsByModifier",
      email
    );
    console.log(
      `Transaction has been evaluated, result is: ${result.toString()}`
    );
    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    if (error == "Error: 404") {
      let response = {
        success: false,
        message: `404 - Currently you have no classics to modify.`,
      };
      res.status(404).json(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});


app.get("/api/Classics/GetByCertifier/", async (req, res) => {
  try {
    const email = req.email;
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.evaluateTransaction(
      "QueryClassicsByCertifier",
      email
    );
    console.log(
      `Transaction has been evaluated, result is: ${result.toString()}`
    );
    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    if (error == "Error: 404") {
      let response = {
        success: false,
        message: `404 - Currently you have no classics to certify.`,
      };
      res.status(404).json(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets the access permissions of a specific classic
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
app.get("/api/Access/Get/:chassisNo", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.evaluateTransaction(
      "GetAccess",
      req.params.chassisNo
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);
    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets the access permissions of a specific classic
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
app.get("/api/Access/CheckLevel/:chassisNo", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const currentDate = new Date();
    const result = await contract.evaluateTransaction(
      "CheckUserAccess",
      req.params.chassisNo,
      currentDate.toISOString(),
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);
    res.status(200).json({type: `${result}`});
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description Gives access permissions to another user. The levels of access can be "modifer", "viewer" or "certifier"
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
app.put("/api/Access/Give/:chassisNo/", async (req, res) => {
  try {
    if (!validNArguments(req, 2)) {
      return errorArguments(res);
    }
    if (!req.body.email) {
      return res.status(400).json(emptyFieldError("'Email to give access'"));
    }
    if (!req.body.level) {
      return res.status(400).json(emptyFieldError("'Level of access'"));
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const currentDate = new Date();
    const result = await contract.submitTransaction(
      "GiveAccess",
      req.params.chassisNo,
      req.body.email,
      req.body.level,
      currentDate.toISOString()
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);
    res.status(200).json({
      success: true,
      message: "Classic's access permissions were updated!",
    });
  } catch (error) {
    if (error.responses[0].response.message === "404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error.responses[0].response.message === "403") {
      forbiddenError(res);
    } else if (error.responses[0].response.message === "404nolevel") {
      let response = {
        success: false,
        message: "404 - level not found",
      };
      console.log("404 - level not found");
      res.status(403).json(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description Revokes access permissions to another user. The levels of access can be "modifer", "viewer" or "certifier"
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
app.put("/api/Access/Revoke/:chassisNo/", async (req, res) => {
  try {
    if (!validNArguments(req, 2)) {
      return errorArguments(res);
    }
    if (!req.body.email) {
      return res.status(400).json(emptyFieldError("'Email to revoke access'"));
    }
    if (!req.body.level) {
      return res.status(400).json(emptyFieldError("'Level of access'"));
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.submitTransaction(
      "RevokeAccess",
      req.params.chassisNo,
      req.body.email,
      req.body.level
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);
    res.status(200).json({
      success: true,
      message: "Classic's access permissions were revoked!",
    });
  } catch (error) {
    if (error.responses[0].response.message === "404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error.responses[0].response.message === "403") {
      forbiddenError(res);
    } else if (error.responses[0].response.message === "404nolevel") {
      let response = {
        success: false,
        message: "404 - level not found",
      };
      console.log("404 - level not found");
      res.status(403).json(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @PUT
 * @async
 * @description A user with certifier access level certifies a classic
 * @returns a sucess or fail message
 * Only the users with certifier access level can execute this function
 */
app.put("/api/Classics/Certify/:chassisNo/", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.submitTransaction(
      "MarkAsCertified",
      req.params.chassisNo
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);
    res.status(200).json({
      success: true,
      message: "Classic was certified!",
    });
  } catch (error) {
    if (error.responses[0].response.message === "404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error.responses[0].response.message === "403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets the history of the access permissions of the classic
 * Only the owner of the classic can execute this function
 */
app.get("/api/Acess/History/:chassisNo", async (req, res) => {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classicCars");
    const result = await contract.evaluateTransaction(
      "GetAccessHistory",
      req.params.chassisNo
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);

    const rr = JSON.parse(result);
    res.status(200).send(rr);
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
});

/**
 * @GET
 * @async
 * @description Gets the info about a Transaction with a given transactionId
 * Only the users with viewer access can perform this function
 */
app.get("/api/Transaction/:chassisNo/:txId", async (req, res) => {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classicCars");
  const currentDate = new Date();
  try {
    access = await contract.evaluateTransaction(
      "ReadClassicAsViewer",
      req.params.chassisNo,
      currentDate.toISOString()
    );
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
  if(Buffer.isBuffer(access)) {
    try {
      const network = await getNetworks(req.email, req.orgName)
      const qsccContract = network.getContract("qscc");
      const result = await qsccContract.evaluateTransaction("GetTransactionByID","mychannel", req.params.txId);
      const decoded = BlockDecoder.decodeTransaction(result);
      const timestamp = decoded.transactionEnvelope.payload.header.channel_header.timestamp;
      const certString = decoded.transactionEnvelope.payload.header.signature_header.creator.id_bytes.toString('utf8');
      const org = decoded.transactionEnvelope.payload.header.signature_header.creator.mspid;
      const cert = new X509Certificate(certString);
      const subjectParts = cert.subject.split('\nCN=');
      const args = decoded.transactionEnvelope.payload.data.actions[0].payload.chaincode_proposal_payload.input.chaincode_spec.input.args
      const decodedArgs = args.map(arg => arg.toString('utf8'));
      const response = decoded.transactionEnvelope.payload.data.actions[0].payload.action.proposal_response_payload.extension.response
      res.status(200).json({
        txId: req.params.txId,
        timestamp: timestamp,
        creator: {email: subjectParts[1], orgMsp: org},
        function: {name: decodedArgs[0], args: decodedArgs.slice(1)},
        response: {response: response.status, result: response.payload.toString('utf8')}
      });
    } catch (error) {
      console.error(error);
      res.status(400).json({error: `${error}`})
    }
  }
});

/**=============> Auxiliar functions <============= */

/**
 * @param {String} field Field that is empty
 * @description Returns response specifying which field is empty
 * @returns response specifying the empty field
 */
function emptyFieldError(field) {
  let response = {
    success: false,
    message:
      field + " field is missing or Invalid in the request. Please try again.",
  };
  return response;
}

/**
 * @param {String} type Type of transaction: evaluate or submit
 * @param {String} error The obtained error
 * @param {Response} res The obtained HTTP Response
 * @description Returns the info about the 500 error to the console e to the HTTP Response
 */
function internalServerError(type, error, res) {
  console.error(`Failed to ${type} transaction: ${error}`);
  res.status(500).json({ error: error });
}

/**
 * @param {String} param The parameter to be printed
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 404 error specifying that the classic was not found
 */
function notFoundError(param, res) {
  let response = {
    success: false,
    message: `404 - The classic ${param} was Not Found`,
  };
  console.log(`404 - The classic ${param} was Not Found`);
  res.status(404).json(response);
}

/**
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 403 error specifying that the user is not authorized
 */
function forbiddenError(res) {
  let response = {
    success: false,
    message: "403 Forbidden - You are not authorized",
  };
  console.log(`403 Forbidden - You are not authorized`);
  res.status(403).json(response);
}

/**
 * @param {Req} req The HTTP Request
 * @param {Integer} number The expected number of arguments
 * Checks if the number of fiels passed in the body of the request is equal to the number of arguments expected
 */
function validNArguments(req, number) {
  const fieldKeys = Object.keys(req.body);
  return fieldKeys.length === number;
}

/**
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 400 error specifying that the user is not authorized
 */
function errorArguments(res) {
  let response = {
    success: false,
    message: `Invalid number of arguments.`,
  };
  res.status(400).json(response);
  console.log("Invalid number of arguments.");
}

/**
 * @param {String} param The parameter to be printed
 * @param {String} error The obtained error
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 409 error specifying that the classic car already exists
 */
function confictErrorClassic(param, error, res) {
  let response = {
    success: false,
    message: `409 - The classic ${param} already exists`,
  };
  console.log(error.responses[0].response.message);
  res.status(409).json(response);
}

/**
 * @param {String} param The parameter to be printed
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 409 error specifying that the user already exists
 */
function confictErrorUser(param, res) {
  let response = {
    success: false,
    message: `User with email ${param} already exists.`,
  };
  console.log(`User with email ${param} already exists.`);
  res.status(409).json(response);
}

/**
 * @param {String} email The email of the user
 * @param {String} orgName The name of the org of the user
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 404 error specifying that the user is not yet registed in the system
 */
function notFoundErrorUser(email, orgName, res) {
  console.log(`User with email ${email} does not exist in the system`);
  res.status(404).json({
    success: false,
    message: `User with email ${email} is not registered with ${orgName}, Please register first.`,
  });
}

/**
 * @async
 * @param {String} email Email of the user
 * @param {String} userOrg  The organization identifier
 * @description Checks if a user is registered or not
 * @returns true if registered, false otherwise
 */
async function isUserRegistered(email, userOrg) {
  let walletUser;
  if (userOrg == "Org1") {
    walletUser = wallet;
  } else if (userOrg == "Org2") {
    walletUser = wallet2;
  } else if (userOrg == "Org3") {
    walletUser = wallet3;
  }
  const userIdentity = await walletUser.get(email);
  if (userIdentity) {
    console.log(userIdentity);
    console.log(`An identity for the user ${email} exists in the wallet`);
    return true;
  }
  return false;
}

/**
 * @async
 * @param {String} email Email of the user
 * @param {String} userOrg  The organization identifier
 * @description Gets the correct gateway for each user
 */
async function getNetworks(email, userOrg) {
  if(gateways[email]) {
    return gateways[email];
  }
  let walletUser;
  let ccpUser;
  if (userOrg == "Org1") {
    walletUser = wallet;
    ccpUser = ccp;
  } else if (userOrg == "Org2") {
    walletUser = wallet2;
    ccpUser = ccp2;
  } else if (userOrg == "Org3") {
    walletUser = wallet3;
    ccpUser = ccp3;
  }
  const gateway = new Gateway();
  const identity = await walletUser.get(email);
  await gateway.connect(ccpUser, {
    walletUser,
    identity,
    discovery: { enabled: true, asLocalhost: true }, // using asLocalhost as this gateway is using a fabric network deployed locally
  });
  const network = await gateway.getNetwork("mychannel");
  gateways[email] = network;
  if (network != {}) {
    return network;
  }
}

async function connectToMongo() {
  const client = new MongoClient(process.env.MONGO_URI);
  try {
    client.connect();
    console.log("Connected to Mongo");
    const database = client.db("jwt-api-local"); // name of database
    usersdb = database.collection("users-local"); // name of collection
  } catch (error) {
    console.log(error);
  }
}

/*  EM STANDBYYYYY
async function registerAdmins(caClient1, wallet1, caClient2, wallet2) {
  await enrollAdmin(caClient1, wallet1, "Org1MSP");
  await registerAndEnrollUser(
    caClient1,
    wallet1,
    "Org1MSP",
    process.env.ID_ADMIN1,
    "org1.department1",
    'admin'
  );
  const hashedpw = await bcrypt.hash(process.env.PASS_ADMIN1, 10);
  usersdb.insertOne({
    email: process.env.ID_ADMIN1,
    password: hashedpw,
    org: "Org1",
    role: "admin",
  });

  await enrollAdmin(caClient2, wallet2, "Org2MSP");
  await registerAndEnrollUser(
    caClient2,
    wallet2,
    "Org2MSP",
    process.env.ID_ADMIN2,
    "org2.department1",
    'admin'
  );
  const hashedpw2 = await bcrypt.hash(process.env.PASS_ADMIN2, 10);
  usersdb.insertOne({
    email: process.env.ID_ADMIN2,
    password: hashedpw2,
    org: "Org2",
    role: "admin",
  });

}*/
