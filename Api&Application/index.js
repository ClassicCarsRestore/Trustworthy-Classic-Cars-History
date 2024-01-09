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
import { buildCCP, buildWallet, sendEmail } from "./auxAPI/AppUtil.js";
import {
  getAllClassics, 
  createClassic,
  updateClassic, 
  updateClassicEmail,
  getClassic,
  addDocument,
  getHistory,
  getClassicsOwner,
  getClassicsModifier,
  getClassicsCertifier,
  certify,
  getTransaction,
} from "./auxAPI/Classics.js";
import {
  createRestorationStep,
  getStep,
  updateStep,
  updateStepPhotos,
  updateStepAndPhotos,
} from "./auxAPI/Restorations.js";
import {
  getAccess,
  giveAcess,
  revokeAccess,
  getAccessHistory,
  checkUserAccess,
} from "./auxAPI/Access.js";
import {
  emptyFieldError,
  internalServerError,
  validNArguments,
  errorArguments,
  confictErrorUser,
  notFoundErrorUser,
  notFoundinSystem,
} from "./auxAPI/AuxFunctions.js";


import { fileURLToPath } from "url";
import { path, dirname } from "path";

//import fs from 'fs'
import fileUpload from "express-fileupload";

//Swagger
import swaggerUi from "swagger-ui-express";
import yaml from "js-yaml";

//CORS 
import cors from "cors";

//HTTPS
import https from "https";

//Fabric Global Variables
let gateways = [];
let wallet;
let wallet2;
let wallet3;
let ccp;

//MONGODB SETTINGS
let usersdb;


// Import the NFTStorage class and File constructor from the 'nft.storage' package
import { NFTStorage} from 'nft.storage';
export let nftstorage;

const app = express();
app.use(bodyParser.json());
app.use(fileUpload());

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const httpsOptions = {
  key: fs.readFileSync(path.join(__dirname, 'cert', 'privkey1.pem')),
  cert: fs.readFileSync(path.join(__dirname, 'cert', 'cert1.pem')),
};

// set secret variable
app.set("secret", process.env.JWT_SECRET);

app.use(bearerToken());

//CORS
app.use(cors());

// Read the Swagger YAML file
const swaggerDocument = yaml.load(fs.readFileSync("./swagger.yaml", "utf8"));

// Validate routes against Swagger documentation
app.use("/api-docs", swaggerUi.serve, swaggerUi.setup(swaggerDocument));

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

const server = https.createServer(httpsOptions, app);

server.listen(8393, async () => {
  try {
    nftstorage = new NFTStorage({ token: process.env.NFT_STORAGE_KEY })

    //Connect to MongoDB
    connectToMongo();

    // build an in memory object with the network configuration (also known as a connection profile)
    ccp = buildCCP();

    // build an instance of the fabric ca services client based on
    // the information in the network configuration
    const caClient = buildCAClient(FabricCAServices, ccp, "org1-ca.default");
    const caClient2 = buildCAClient(FabricCAServices, ccp, "org2-ca.default");
    const caClient3 = buildCAClient(FabricCAServices, ccp, "org3-ca.default");

    // setup the wallet to hold the credentials of the application user
    wallet = await buildWallet(Wallets, path.join(__dirname, "org1-wallet"));
    wallet2 = await buildWallet(Wallets, path.join(__dirname, "org2-wallet"));
    wallet3 = await buildWallet(Wallets, path.join(__dirname, "org3-wallet"));

    await enrollAdmin(caClient, wallet, "Org1MSP");
    await enrollAdmin(caClient2, wallet2, "Org2MSP");
    await enrollAdmin(caClient3, wallet3, "Org3MSP");

    await registerAndEnrollUser(
      caClient,
      wallet,
      "Org1MSP",
      "appUser",
      "org1.department1",
      "client"
    );

    // Create a new gateway instance for interacting with the fabric network.
    const gateway = new Gateway();

    try {
      // Setup the gateway instance
      // The user will now be able to create connections to the fabric network and be able to
      // submit transactions and query.
      await gateway.connect(ccp, {
        wallet,
        identity: "appUser",
        discovery: { enabled: true, asLocalhost: false },
      });

      // Build a network instance based on the channel where the smart contract is deployed
      const network = await gateway.getNetwork("mychannel");

      // Get the contract from the network.
      const contract = network.getContract("classiccars");

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
  if (!validNArguments(req, 6)) {
    return errorArguments(res);
  }
  let email = req.body.email;
  let password = req.body.password;
  let orgName = req.body.orgname;
  let firstname = req.body.firstname;
  let surname = req.body.surname;
  let country = req.body.country;
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
  if (!country) {
    return res.status(400).json(emptyFieldError("'country'"));
  }
  let userInWallet = await isUserRegistered(email, orgName);
  if (userInWallet) {
    return confictErrorUser(email, res);
  }
  try {
    let orgMspIdUser;
    let caHostNameUser;
    let userWallet;
    let affiliationUser;
    let role;
    if (orgName == "Org1") {
      orgMspIdUser = "Org1MSP";
      caHostNameUser = "org1-ca.default";
      userWallet = wallet;
      role = "client";
    } else if (orgName == "Org2") {
      orgMspIdUser = "Org2MSP";
      caHostNameUser = "org2-ca.default";
      userWallet = wallet2;
      role = "workshop";
    } else if (orgName == "Org3") {
      orgMspIdUser = "Org3MSP";
      caHostNameUser = "org3-ca.default";
      userWallet = wallet3;
      role = "certifier";
    }

    const caClientUser = buildCAClient(FabricCAServices, ccp, caHostNameUser);
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
      surname: surname,
      country: country
    });

    let result = {
      email: email,
      success: true,
    };

    sendEmail(email, "ClassicsChain - New Account", "A new account has been registered with your email.");

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

  const dbUser = await usersdb.findOne({ email: email });
  if (!dbUser) {
    return notFoundErrorUser(email, orgName, res);
  }
  const userAllowed = await bcrypt.compare(password, dbUser.password);
  let userInWallet = await isUserRegistered(dbUser.email, orgName);

  if (userInWallet && userAllowed) {
    let token = jwt.sign(
      {
        exp: Math.floor(Date.now() / 1000) + 18000,
        email: dbUser.email,
        orgName: dbUser.org,
      },
      app.get("secret")
    );

    res.json({
      success: true,
      message: { email: dbUser.email, token: token, org: dbUser.org},
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
 * @GET
 * @async
 * @description Auxiliar function to get all the users registered in the system
 * @returns a sucess or fail message
 */
app.get("/api/Users/GetAll", async (req, res) => {
  try {
    const projection = { projection: { _id: 0, firstname: 1, surname:1, country:1, email: 1, role: 1 } };
    const users = await usersdb.find({}, projection).toArray();
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
    const projection = { projection: { _id: 0, firstname: 1, surname:1, country:1, email: 1, role: 1 } };
    const dbUser = await usersdb.findOne({ email: req.params.email }, projection);
    if (!dbUser) {
      notFoundinSystem(email, res);
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
  getAllClassics(req,res);
});

/**
 * @POST
 * @async
 * @description Creates a new classic with details got from the request body, and adds this classic
 * to the ledger
 */
app.post("/api/Classics/Create", async (req, res) => {
  createClassic(req, res, usersdb);
});

/**
 * @PUT
 * @async
 * @description Updates the allowed details of a classic.
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.put("/api/Classics/Update/:chassisNo", async (req, res) => {
  updateClassic(req, res);
});

/**
 * @PUT
 * @async
 * @description Changes the email of the classic's owner (works as ownership change);
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.put("/api/Classics/UpdateEmail/:chassisNo/:newEmail", async (req, res) => {
  updateClassicEmail(req, res, usersdb)
});

/**
 * @PUT
 * @async
 * @description Updates the classic's documentation by adding a new restoration procedure with details
 * got from the request body, including any type of file
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.post("/api/Restorations/Create/:chassisNo", async (req, res) => {
  createRestorationStep(req, res);
});

/**
 * @GET
 * @async
 * @description Gets a specific restoration procedure of a classic
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.get("/api/Restorations/Get/:chassisNo/:stepId", async (req, res) => {
  getStep(req, res);
});

/**
 * @PUT
 * @async
 * @description Updates a specific restoration procedure of a classic
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.put("/api/Restorations/Update/:chassisNo", async (req, res) => {
  updateStep(req, res);
});

/**
 * @PUT
 * @async
 * @description Updates a specific restoration procedure of a classic with new photos/documents
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.put("/api/Restorations/Update/:chassisNo/Photos", async (req, res) => {
  updateStepPhotos(req, res);
});

/**
 * @PUT
 * @async
 * @description Updates a specific restoration procedure of a classic with new title, description and photos/documents
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
app.put("/api/Restorations/UpdateAndPhotos/:chassisNo", async (req, res) => {
  updateStepAndPhotos(req, res);
});

/**
 * @GET
 * @async
 * @description Gets the details of a specified classic with the chassisNo got form the url parameters
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.get("/api/Classics/Get/:chassisNo", async (req, res) => {
  getClassic(req, res);
});

/**	
 * @PUT	
 * @async	
 * @description Updates a specific classic with new documentation	
 * @returns a sucess or fail message	
 * Only the users with modifier access level can execute this function	
 */
app.put("/api/Classics/AddDocument/:chassisNo", async (req, res) => {
  addDocument(req, res);
});

/**
 * @GET
 * @async
 * @description Gets the history of a specified classic with the chassisNo got form the url parameters
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
app.get("/api/Classics/History/:chassisNo", async (req, res) => {
  getHistory(req, res);
});

/**	
 * @GET	
 * @async	
 * @description Gets all the classics owned by a specific user	
 * @returns List of classics if the user making the request ownes at least one classic, error otherwise	
 */	
app.get("/api/Classics/GetByOwner/", async (req, res) => {	
  getClassicsOwner(req, res);
});

/**	
 * @GET	
 * @async	
 * @description Gets all the cars that a workshop/modifier email has access to
 * @returns List of all the cars that a workshop/modifier email has access to, error otherwise
 */	
app.get("/api/Classics/GetByModifier/", async (req, res) => {	
  getClassicsModifier(req, res);
});

/**	
 * @GET	
 * @async	
 * @description Gets all the cars that a certifier entity email has access to
 * @returns List of all the cars that a certifier entity email has access to, error otherwise
 */	
app.get("/api/Classics/GetByCertifier/", async (req, res) => {	
  getClassicsCertifier(req, res);
});

/**	
 * @GET	
 * @async	
 * @description Gets the access permissions of a specific classic	
 * @returns a sucess or fail message	
 * Only the owner of the classic can execute this function	
 */
app.get("/api/Access/Get/:chassisNo", async (req, res) => {
  getAccess(req, res);
});

/**	
 * @PUT	
 * @async	
 * @description Gives access permissions to another user. The levels of access can be "modifer", "viewer" or "certifier"	
 * @returns a sucess or fail message	
 * Only the owner of the classic can execute this function	
 */
app.put("/api/Access/Give/:chassisNo/", async (req, res) => {
  giveAcess(req, res, usersdb);
});

/**	
 * @PUT	
 * @async	
 * @description Revokes access permissions to another user. The levels of access can be "modifer", "viewer" or "certifier"	
 * @returns a sucess or fail message	
 * Only the owner of the classic can execute this function	
 */
app.put("/api/Access/Revoke/:chassisNo/", async (req, res) => {
  revokeAccess(req, res, usersdb);
});

/**	
 * @PUT	
 * @async	
 * @description A user with certifier access level certifies a classic	
 * @returns a sucess or fail message	
 * Only the users with certifier access level can execute this function	
 */
app.put("/api/Classics/Certify/:chassisNo/", async (req, res) => {
  certify(req, res);
});

/**
 * @GET
 * @async
 * @description Gets the access permissions of a specific classic
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
app.get("/api/Access/CheckLevel/:chassisNo", async (req, res) => {
  checkUserAccess(req, res);
});

/**	
 * @GET	
 * @async	
 * @description Gets the history of the access permissions of the classic	
 * Only the owner of the classic can execute this function	
 */
app.get("/api/Acess/History/:chassisNo", async (req, res) => {
  getAccessHistory(req, res);
});

/**
 * @GET
 * @async
 * @description Gets the info about a Transaction with a given transactionId
 * Only the users with viewer access can perform this function
 */
app.get("/api/Transaction/:chassisNo/:txId", async (req, res) => {
  getTransaction(req, res);
});

/**=============> Auxiliar functions <============= */

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
export async function getNetworks(email, userOrg) {
  if(gateways[email]) {
    return gateways[email];
  }
  let walletUser;
  if (userOrg == "Org1") {
    walletUser = wallet;
  } else if (userOrg == "Org2") {
    walletUser = wallet2;
  } else if (userOrg == "Org3") {
    walletUser = wallet3;
  }
  const gateway = new Gateway();
  const identity = await walletUser.get(email);
  await gateway.connect(ccp, {
    walletUser,
    identity,
    discovery: { enabled: true, asLocalhost: false },
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
    const database = client.db("jwt-api"); // name of database
    usersdb = database.collection("users"); // name of collection
  } catch (error) {
    console.log(error);
  }
}
