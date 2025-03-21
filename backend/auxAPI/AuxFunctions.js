'use strict';

/**
 * @param {Response} res The obtained HTTP Response
 * @param {Buffer} result The result to be printed
 * @description Returns a 404 error specifying that the classic was not found
 */
export function evaluateSuccess(res, result) {
    console.log(`Transaction has been evaluated, result is: ${result.toString()}`);
    const rr = JSON.parse(result);
    res.status(200).send(rr);
}

/**
 * @param {String} field Field that is empty
 * @description Returns response specifying which field is empty
 * @returns response specifying the empty field
 */
export function emptyFieldError(field) {
    let response = {success: false, message: field + " field is missing or Invalid in the request. Please try again.",};
    return response;
}

/**
 * @param {String} type Type of transaction: evaluate or submit
 * @param {String} error The obtained error
 * @param {Response} res The obtained HTTP Response
 * @description Returns the info about the 500 error to the console e to the HTTP Response
 */
export function internalServerError(type, error, res) {
    res.status(500).json({ error: error });
    console.error(`Failed to ${type} transaction: ${error}`);
}

/**
 * @param {String} param The parameter to be printed
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 404 error specifying that the classic was not found
 */
export function notFoundError(param, res) {
    let response = { success: false, message: `404 - The classic ${param} was Not Found`,};
    res.status(404).json(response);
    console.log(`404 - The classic ${param} was Not Found`);
}
  
/**
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 403 error specifying that the user is not authorized
 */
export function forbiddenError(res) {
    let response = {success: false, message: "403 Forbidden - You are not authorized",};
    res.status(403).json(response);
    console.log(`403 Forbidden - You are not authorized`);
}

/**
 * @param {Req} req The HTTP Request
 * @param {Integer} number The expected number of arguments
 * Checks if the number of fiels passed in the body of the request is equal to the number of arguments expected
 */
export function validNArguments(req, number) {
    const fieldKeys = Object.keys(req.body);
    return fieldKeys.length === number;
}

/**
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 400 error specifying that the user is not authorized
 */
export function errorArguments(res) {
    let response = {success: false,message: `Invalid number of arguments.`,};
    res.status(400).json(response);
    console.log("Invalid number of arguments.");
}

/**
 * @param {String} param The parameter to be printed
 * @param {String} error The obtained error
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 409 error specifying that the classic car already exists
 */
export function confictErrorClassic(param, error, res) {
    let response = { success: false, message: `409 - The classic ${param} already exists`,};
    res.status(409).json(response);
    console.log(error.responses[0].response.message);
}

/**
 * @param {String} param The parameter to be printed
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 409 error specifying that the user already exists
 */
export function confictErrorUser(param, res) {
    let response = { success: false, message: `User with email ${param} already exists.`,};
    res.status(409).json(response);
    console.log(`User with email ${param} already exists.`);
}

/**
 * @param {String} email The email of the user
 * @param {String} orgName The name of the org of the user
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 404 error specifying that the user is not yet registed in the system
 */
export function notFoundErrorUser(email, orgName, res) {
    let response = {success: false, message: `User with email ${email} is not registered with ${orgName}, Please register first.`,};
    res.status(404).json(response);
    console.log(`User with email ${email} does not exist in the system`);
}

/**
 * @param {String} email The email of the user
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 404 error specifying that the user is not yet registed in the system (developing)
 */
export function notFoundinSystem(email, res) {
    let response = {success: false, message: `User with email ${email} is not registered in the system.`,};
    res.status(404).json(response);
    console.log(`User with email ${email} is not registered in the system.`);
}

/**
 * @param {String} email The email of the user
 * @param {Response} res The obtained HTTP Response
 * @description Returns a 404 error specifying that the user is not registed in the system
 */
export function otherUserError(email, res) {
    let response = {success: false, message: `User with email ${email} is not registered in the system.`,};
    res.status(404).json(response);
    console.log(`User with email ${email} does not exist in the system`);
}

export function noStepError(res, stepId) {
    const response = `The step ${stepId} was not found.`;
    res.status(404).json({success: false, message: response,});
    console.log(response);
}

export function stepNotMadeBy(res, stepId) {
    const response = `Not authorized to update step ${stepId} since you are who made it originally.`;
    res.status(404).json({success: false, message: response,});
    console.log(response);
}

export function noLevelError(res) {
    const response = "The provided level is not correct. Valid values: viewer, certifier, modifier";	
    res.status(404).json({success: false, message: response,});	
    console.log(response);	
}