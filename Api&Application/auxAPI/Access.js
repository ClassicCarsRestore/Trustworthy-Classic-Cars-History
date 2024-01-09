"use strict";
import {
  evaluateSuccess,
  internalServerError,
  validNArguments,
  emptyFieldError,
  notFoundError,
  forbiddenError,
  otherUserError,
  errorArguments,
  noLevelError,
} from "./AuxFunctions.js";
import { getNetworks } from "../index.js";
import {sendEmail} from "./AppUtil.js";

/**
 * @export
 * @async
 * @description Gets the access permissions of a specific classic
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
export async function getAccess(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const result = await contract.evaluateTransaction(
      "GetAccess",
      req.params.chassisNo
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
 * @description Gives access permissions to another user. The levels of access can be "modifer", "viewer" or "certifier"
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
export async function giveAcess(req, res, usersdb) {
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
    const dbUser = await usersdb.findOne({ email: req.body.email });
    if (!dbUser) {
      return otherUserError(req.body.email, res);
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const currentDate = new Date();
    await contract.submitTransaction(
      "GiveAccess",
      req.params.chassisNo,
      req.body.email,
      req.body.level,
      currentDate.toISOString() 
    );
    const response = "Classic's access permissions were updated!";
    sendEmail(req.body.email, "ClassicsChain - New Access granted", "You have been granted new "+req.body.level +" access permissions to the classic with VIN "+req.params.chassisNo+".\n Please access it in this link: http://194.210.120.34:4200/classics/details/"+req.params.chassisNo);
    res.status(200).json({ success: true, message: response });
    console.log(response);
  } catch (error) {
    if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else if (error.responses[0].response.message === "404nolevel") {
        noLevelError(res);	
      } 
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Revokes access permissions to another user. The levels of access can be "modifer", "viewer" or "certifier"
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
export async function revokeAccess(req, res, usersdb) {
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
    const dbUser = await usersdb.findOne({ email: req.body.email });
    if (!dbUser) {
      return otherUserError(req.body.email, res);
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    await contract.submitTransaction("RevokeAccess", req.params.chassisNo, req.body.email, req.body.level);
    const response = "Classic's access permissions were revoked!";
    res.status(200).json({ success: true, message: response });
    console.log(response);
  } catch (error) {
    if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else if (error.responses[0].response.message === "404nolevel") {
        noLevelError(res);	
      } 
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Gets the access permissions of a specific classic
 * @returns a sucess or fail message
 * Only the owner of the classic can execute this function
 */
export async function checkUserAccess(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const currentDate = new Date();
    const result = await contract.evaluateTransaction(
      "CheckUserAccess",
      req.params.chassisNo,
      currentDate.toISOString()
    );
    console.log(`Transaction has been evaluated, result is: ${result}`);
    res.status(200).json({ type: `${result}` });
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
 * @description Gets the history of the access permissions of the classic
 * Only the owner of the classic can execute this function
 */
export async function getAccessHistory(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const result = await contract.evaluateTransaction(
      "GetAccessHistory",
      req.params.chassisNo
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
