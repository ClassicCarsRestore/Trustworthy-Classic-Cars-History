"use strict";
import {
  evaluateSuccess,
  internalServerError,
  validNArguments,
  emptyFieldError,
  notFoundError,
  forbiddenError,
  noStepError,
  stepNotMadeBy,
  errorArguments
} from "./AuxFunctions.js";
import {nftstorage, getNetworks } from "../index.js";
import {packToBlob} from 'ipfs-car/pack/blob'

/**
 * @export
 * @async
 * @description Updates the classic's documentation by adding a new restoration procedure with details
 * got from the request body, including any type of file
 * Only the user with the email equal to the one registered in the classic can perform this action
 */
export async function createRestorationStep(req, res) {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classiccars");
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
        const { root, car } = await packToBlob({
          input: [{
              path: req.params.chassisNo,
              content: file.data
          }]
        });
        const cid = `https://nftstorage.link/ipfs/${root}/${req.params.chassisNo}`
        nftstorage.storeCar(car);
        console.log(cid);
        photosIds.push(cid);
        console.log("File was uploaded!");
      } else {
        console.log("Uploading files...");
        for (const file of req.files.file) {
          const { root, car } = await packToBlob({
            input: [{
                path: req.params.chassisNo,
                content: file.data
            }]
          });
          const cid = `https://nftstorage.link/ipfs/${root}/${req.params.chassisNo}`
          nftstorage.storeCar(car);
          console.log(cid);
          photosIds.push(cid);
        }
        console.log("Files were uploaded!");
      }
      stringified = JSON.stringify(photosIds);
      const currentDate = new Date();
      const afterStepId = await contract.submitTransaction(
        "CreateRestorationStep", 
        req.params.chassisNo,
        req.body.title,
        req.body.description,
        stringified,
        currentDate.toISOString()
      );
      const response = `The car with chassis number ${req.params.chassisNo} was updated with a new restoration procedure.`;
      res.status(200).json({ success: true, message: response, stepId: afterStepId.toString('utf8') });
      console.log(response);
      console.log(afterStepId.toString('utf8'));
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
 * @description Gets a specific restoration procedure of a classic
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
export async function getStep(req, res) {
  try {
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    const result = await contract.evaluateTransaction(
      "GetStep",
      req.params.chassisNo,
      req.params.stepId
    );
    evaluateSuccess(res, result);
  } catch (error) {
    if (error == "Error: 404") {
      notFoundError(req.params.chassisNo, res);
    } else if (error == "Error: 403") {
      forbiddenError(res);
    } else if (error == "Error: 404nostep") {
      noStepError(res, req.params.stepId);
    } else {
      internalServerError("evaluate", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Updates a specific restoration procedure of a classic
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
export async function updateStep(req, res) {
  try {
    if (!validNArguments(req, 3)) {
      return errorArguments(res);
    }
    const network = await getNetworks(req.email, req.orgName)
    const contract = network.getContract("classiccars");
    await contract.submitTransaction(
      "UpdateStep",
      req.params.chassisNo,
      req.body.stepId,
      req.body.newTitle,
      req.body.newDescription
    );
    const response = `The step ${req.body.stepId} was updated sucessfully.`;
    res.status(200).json({ success: true, message: response });
    console.log(response);
  } catch (error) {
    if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else if (error.responses[0].response.message === "404madeby") {
        stepNotMadeBy(res, req.body.stepId);
      } else if (error.responses[0].response.message === "404nostep") {
        noStepError(res, req.body.stepId);
      } 
    }else {
      internalServerError("submit", error, res);
    }
  }
}

/**
 * @export
 * @async
 * @description Updates a specific restoration procedure of a classic with new photos/documents
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
export async function updateStepPhotos(req, res) {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classiccars");
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
        return res
          .status(400)
          .json({ success: false, message: "Please add at least one file!" });
      } else if (!Array.isArray(req.files.file)) {
        console.log("Uploading a file...");
        const file = req.files.file;
        const { root, car } = await packToBlob({
          input: [{
              path: req.params.chassisNo,
              content: file.data
          }]
        });
        const cid = `https://nftstorage.link/ipfs/${root}/${req.params.chassisNo}`
        nftstorage.storeCar(car);
        console.log(cid);
        photosIds.push(cid);
        console.log("File was uploaded!");
      } else {
        console.log("Uploading files...");
        for (const file of req.files.file) {
          const { root, car } = await packToBlob({
            input: [{
                path: req.params.chassisNo,
                content: file.data
            }]
          });
          const cid = `https://nftstorage.link/ipfs/${root}/${req.params.chassisNo}`
          nftstorage.storeCar(car);
          console.log(cid);
          photosIds.push(cid);
        }
        console.log("Files were uploaded!");
      }
      stringified = JSON.stringify(photosIds);
      await contract.submitTransaction("UpdateStepPhotos", req.params.chassisNo, req.body.stepId, stringified);
      const response = `The photos/documents of step ${req.body.stepId} were updated sucessfully.`;
      res.status(200).json({ success: true, message: response });
      console.log(response);
    } catch (error) {
      if (error.responses && error.responses.length > 0 && error.responses[0].hasOwnProperty("response")) {
        if (error.responses[0].response.message === "404") {
          notFoundError(req.params.chassisNo, res);
        } else if (error.responses[0].response.message === "403") {
          forbiddenError(res);
        } else if (error.responses[0].response.message === "404madeby") {
          stepNotMadeBy(res, req.body.stepId);
        } else if (error.responses[0].response.message === "404nostep") {
          noStepError(res, req.body.stepId);
        } 
      }else {
        internalServerError("submit", error, res);
      }
    }
  }
}

/**
 * @export
 * @async
 * @description Updates a specific restoration procedure of a classic with new title, description and photos/documents
 * @returns a sucess or fail message
 * Only the users with modifier access level can execute this function
 */
export async function updateStepAndPhotos(req, res) {
  let access;
  const network = await getNetworks(req.email, req.orgName)
  const contract = network.getContract("classiccars");
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
        const { root, car } = await packToBlob({
          input: [{
              path: req.params.chassisNo,
              content: file.data
          }]
        });
        const cid = `https://nftstorage.link/ipfs/${root}/${req.params.chassisNo}`
        nftstorage.storeCar(car);
        console.log(cid);
        photosIds.push(cid);
        console.log("File was uploaded!");
      } else {
        console.log("Uploading files...");
        for (const file of req.files.file) {
          const { root, car } = await packToBlob({
            input: [{
                path: req.params.chassisNo,
                content: file.data
            }]
          });
          const cid = `https://nftstorage.link/ipfs/${root}/${req.params.chassisNo}`
          nftstorage.storeCar(car);
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
        message: `The title, description and photos/documents of step ${req.body.stepId} were updated sucessfully.`,
      });
      console.log(`The restoration ${req.body.stepId} was updated.`);
    } catch (error) {
      if (error.responses[0].response.message === "404") {
        notFoundError(req.params.chassisNo, res);
      } else if (error.responses[0].response.message === "403") {
        forbiddenError(res);
      } else if (error.responses[0].response.message === "404madeby") {
        stepNotMadeBy(res, req.body.stepId);
      } else if (error.responses[0].response.message === "404nostep") {
        noStepError(res, req.body.stepId);
      } else {
        internalServerError("submit", error, res);
      }
    }
  }
}
