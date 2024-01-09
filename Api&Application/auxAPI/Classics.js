"use strict";
import {
  evaluateSuccess,
  internalServerError,
  otherUserError,
  confictErrorClassic,
  validNArguments,
  emptyFieldError,
  notFoundError,
  forbiddenError,
  errorArguments,
} from "./AuxFunctions.js";
import {nftstorage, getNetworks } from "../index.js";
import {sendEmail} from "./AppUtil.js";
//Transaction
import { BlockDecoder } from 'fabric-common';
import { X509Certificate } from 'crypto';
import {packToBlob} from 'ipfs-car/pack/blob'

/**
 * @export
 * @async
 * @description Gets all the available classics in the system (works as homepage for admins)
 */
export async function getAllClassics(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const result = await contract.evaluateTransaction("GetAllClassics");
    evaluateSuccess(res, result);
  } catch (error) {
    internalServerError("evaluate", error, res);
  }
}

/**
 * @export
 * @async
 * @description Creates a new classic with details got from the request body, and adds this classic
 * to the ledger
 */
export async function createClassic(req, res, usersdb) {
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
    const dbUser = await usersdb.findOne({ email: req.body.ownerEmail });
    if (!dbUser) {
      return otherUserError(req.body.ownerEmail, res);
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const currentDate = new Date();
    if(req.orgName == "Org3") {
      const endorsingPeers = ["org1-peer0.default", "org2-peer0.default"];
      let tx = contract.createTransaction("CreateClassic");
      tx.setEndorsingPeers(endorsingPeers);
      await tx.submit(req.body.make, req.body.model, req.body.year, req.body.licencePlate, req.body.country, req.body.chassisNo, req.body.engineNo, req.body.ownerEmail, currentDate.toISOString());
    } 
    else {
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
    }
    const response = `The car with chassis number ${req.body.chassisNo} was added successfully!`;
    console.log(response);
    sendEmail(req.body.ownerEmail, "ClassicsChain - New Classic registered to your account", "A new classic with VIN "+ req.body.chassisNo +" was registered to your account by " +req.email +".");
    res.status(200).json({ success: true, message: response });
  } catch (error) {
    if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
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
}

/**
 * @export
 * @async
 * @description Updates the allowed details of a classic. Only owner can use this function
 */
export async function updateClassic(req, res) {
  try {
    if (!validNArguments(req, 6)) {
      return errorArguments(res);
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    await contract.submitTransaction(
      "UpdateClassic",
      req.params.chassisNo, req.body.make, req.body.model, req.body.year, req.body.licencePlate, req.body.country, req.body.engineNo
    );
    const response = `The car with chassis number ${req.params.chassisNo} was updated sucessfully.`;
    sendEmail(req.email, "ClassicsChain - Your Classic's details were updated", "The classic with VIN "+ req.params.chassisNo +" had some of its details modified.");
    res.status(200).json({ success: true, message: response });
    console.log(response);
  } catch (error) {
    if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      }
    } else {
      internalServerError("submit", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Changes the email of the classic's owner (works as ownership change);
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
export async function updateClassicEmail(req, res, usersdb) {
  try {
    const dbUser = await usersdb.findOne({ email: req.params.newEmail });
    if (!dbUser) {
      return otherUserError(req.params.newEmail, res);
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    await contract.submitTransaction("UpdateClassicEmail", req.params.chassisNo, req.params.newEmail);
    const response = `The car with chassis number ${req.params.chassisNo} was updated to the email ${req.params.newEmail}.`;
    sendEmail(req.params.newEmail, "ClassicsChain - New Classic transfered to your account", "A classic with VIN "+ req.params.chassisNo +" was transfered to your account by " +req.email +".");
    sendEmail(req.email, "ClassicsChain - Your Classic was transfered", "The ownership of your classic with VIN "+ req.params.chassisNo +" was transfered to " +req.params.newEmail +".");
    res.status(200).json({ success: true, message: response });
    console.log(response);
  } catch (error) {
    if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } 
    } else {
      internalServerError("submit", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Gets the details of a specified classic with the chassisNo got form the url parameters
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
export async function getClassic(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const currentDate = new Date();
    const result = await contract.evaluateTransaction(
      "ReadClassicAsViewer",
      req.params.chassisNo,
      currentDate.toISOString()
    );
    evaluateSuccess(res, result);
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Updates a specific classic with new documentation
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
export async function addDocument(req, res) {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classiccars");
  try {
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
      console.log("Uploading document...");
      const filename=req.body.name
      let doc = req.files.file;
      const { root, car } = await packToBlob({
        input: [{
            path: filename,
            content: doc.data
        }]
      });
      const docId = `https://nftstorage.link/ipfs/${root}/${filename}`
      nftstorage.storeCar(car);
      console.log("Document was uploaded!");
      if(req.orgName == "Org3") {
        const endorsingPeers = ["org1-peer0.default", "org2-peer0.default"];
        let tx = contract.createTransaction("AddDocument");
        tx.setEndorsingPeers(endorsingPeers);
        await tx.submit(req.params.chassisNo, req.body.name, docId);
      } 
      else {
        await contract.submitTransaction("AddDocument", req.params.chassisNo, req.body.name, docId);
      }
      const response = `A new document was added to the car with chassis number ${req.params.chassisNo}.`;
      res.status(200).json({ success: true, message: response });
      console.log(response);
    } catch (error) {
      if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
        if (error.responses[0].response.message === "404") {
          notFoundError(req.params.chassisNo, res);
        } else if (error.responses[0].response.message === "403") {
          forbiddenError(res);
        } 
      } else {
        internalServerError("submit", error, res);
      }
    }
  }
}

/**
 * @export
 * @async
 * @description Gets the history of a specified classic with the chassisNo got form the url parameters
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
export async function getHistory(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const currentDate = new Date();
    const result = await contract.evaluateTransaction(
      "GetClassicHistory2",
      req.params.chassisNo,
      currentDate.toISOString()
    );
    evaluateSuccess(res, result);
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Gets all the classics owned by a specific user
 * @returns List of classics if the user making the request ownes at least one classic, error otherwise
 */
export async function getClassicsOwner(req, res) {
  try {
    const email = req.email;
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const result = await contract.evaluateTransaction(
      "QueryClassicsByOwner",
      email
    );
    evaluateSuccess(res, result);
  } catch (error) {
    if (error == "Error: 404") {
      const response = `404 - User ${req.email} has no classics registered in the system`;
      res.status(404).json({ success: false, message: response });
      console.log(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Gets all the cars that a workshop/modifier email has access to
 * @returns List of all the cars that a workshop/modifier email has access to, error otherwise
 */
export async function getClassicsModifier(req, res) {
  try {
    const email = req.email;
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const result = await contract.evaluateTransaction(
      "QueryClassicsByModifier",
      email
    );
    evaluateSuccess(res, result);
  } catch (error) {
    if (error == "Error: 404") {
      let response = {success: false,message: `404 - Currently you have no classics to modify.`,};
      res.status(404).json(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Gets all the cars that a certifier entity email has access to
 * @returns List of all the cars that a certifier entity email has access to, error otherwise
 */
export async function getClassicsCertifier(req, res) {
  try {
    const email = req.email;
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const result = await contract.evaluateTransaction(
      "QueryClassicsByCertifier",
      email
    );
    evaluateSuccess(res, result);
  } catch (error) {
    if (error == "Error: 404") {
      let response = { success: false, message: `404 - Currently you have no classics to certify.`,};
      res.status(404).json(response);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**	
 * @export	
 * @async	
 * @description A user with certifier access level certifies a classic	
 * @returns a sucess or fail message	
 * Only the users with certifier access level can execute this function	
 */
export async function certify(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const endorsingPeers = ["org1-peer0.default", "org2-peer0.default"];
    let tx = contract.createTransaction("MarkAsCertified");
    tx.setEndorsingPeers(endorsingPeers);
    await tx.submit(req.params.chassisNo);
    const response = "Classic was certified!";
    res.status(200).json({ success: true, message: response });
    console.log(response);
  } catch (error) {
    if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } 
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Gets the info about a Transaction with a given transactionId
 * @returns a success or fail message
 * Only the users with viewer access can perform this function
 */
export async function getTransaction(req, res) {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classiccars");
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
      //const network = await getNetworks(req.email, req.orgName)
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
        fx: {name: decodedArgs[0], args: decodedArgs.slice(1)},
        response: {response: response.status, result: response.payload.toString('utf8')}
      });
    } catch (error) {
      console.error(error);
      res.status(400).json({error: `${error}`})
    }
  }
}
