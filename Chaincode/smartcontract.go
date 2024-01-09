package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	"encoding/base64"
	"time"
	"strings"
)

type ServerConfig struct {
	CCID    string
	Address string
}

// SmartContract provides functions for managing a Classic Car Asset
type SmartContract struct {
	contractapi.Contract
}

// RestorationStep struct describes basic details of a restoration procedure
type RestorationStep struct {
	ID 			string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PhotosIds   []string `json:"photosIds"`
	MadeBy		string   `json:"madeBy"`
	When 		string   `json:"when"`
}


// Classic struct describes basic details of what makes up a classic, chassisNo is the main key
type Classic struct {
	Make				string `json:"make"`
	Model				string `json:"model"`
	Year				int    `json:"year"`
	LicencePlate		string `json:"licencePlate"`
	Country				string `json:"country"`
	ChassisNo			string `json:"chassisNo"`
	EngineNo			string `json:"engineNo"`
	OwnerEmail			string `json:"ownerEmail"`
	PriorOwners			[]string `json:"priorOwners"`
	Restorations  		[]RestorationStep `json:"restorations"`
	Certifications   	[]string `json:"certifications"`
	Documents 			map[string]string `json:"documents"`
}

// Access struct describes access permission details
type Access struct {
	OwnerEmail			string `json:"ownerEmail"`
	Viewers				map[string]string `json:"viewers"`
	Modifiers			map[string]string `json:"modifiers"`
	Certifiers 			map[string]string `json:"certifiers"`
}


// InitLedger initiates the ledger, adding a base classic to it
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	priorOwners := []string{}
	restorations := []RestorationStep{}
	certifications := []string{}
	documents := map[string]string{}
	classics := []Classic{
		{Make: "Jaguar", Model: "E", Year: 1964, LicencePlate: "AWA 69B", Country: "Portugal", ChassisNo: "07642", EngineNo: "ER8595", OwnerEmail: "raimundo.branco@yopmail.com", PriorOwners: priorOwners, Restorations: restorations, Certifications: certifications, Documents: documents},
	}

	for _, classic := range classics {
		classicJSON, err := json.Marshal(classic)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState("Classic_"+classic.ChassisNo, classicJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}

	}

	PrintCreatorInfo(ctx)

	return nil
}

// GetAllClassics returns all classics found in world state
func (s *SmartContract) GetAllClassics(ctx contractapi.TransactionContextInterface) ([]*Classic, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("Classic_", "Classic_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var classics []*Classic
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var classic Classic
		err = json.Unmarshal(queryResponse.Value, &classic)
		if err != nil {
			return nil, err
		}
		classics = append(classics, &classic)
	}

	return classics, nil
}


/*
* @param Transaction Context
* @param {String} make The make of the classic
* @param {String} model The model of the classic
* @param {Integer} year The year of manufacturing the classic
* @param {String} licencePlate The license plate of manufacturing the classic
* @param {String} country The registration country of the classic
* @param {String} engineNo The engine number of the classic
* @param {String} chassisNo The chassis number of the classic
* @param {String} ownerEmail The email of the owner
* @description CreateClassic issues a new asset to the world state with given details. (Only a verified workshop can perform this function)
* @returns The classic stored in the world state
*/
func (s *SmartContract) CreateClassic(ctx contractapi.TransactionContextInterface, make string, model string, year int, licencePlate string, country string, chassisNo string, engineNo string, ownerEmail string, currentTime string) error {
	org, ok, error := ctx.GetClientIdentity().GetAttributeValue("org")
	if error != nil {
		return error
	}
	if !ok {
		return error
	}
	if (org != "Org2MSP" && org != "Org3MSP") {
		return fmt.Errorf("403")
	}
	priorOwners := []string{}
	restorations := []RestorationStep{}
	certifications := []string{}
	documents := map[string]string{}
	//===
	viewers := map[string]string{}
	modifiers := map[string]string{}
	certifiers := map[string]string{}
	//====
	exists, err := s.ClassicExists(ctx, chassisNo)
	if err != nil {
		return err
	}
	if exists {
		//The classic with chassis number chassisNo already exists
		return fmt.Errorf("409")
	}

	classic := Classic{
		Make: make,
		Model: model,		
		Year: year,
		LicencePlate: licencePlate,
		Country: country,
		ChassisNo: chassisNo,
		EngineNo: engineNo,
		OwnerEmail: ownerEmail,
		PriorOwners: priorOwners,
		Restorations: restorations,
		Certifications: certifications,
		Documents: documents,
	}
	classicJSON, err := json.Marshal(classic)
	if err != nil {
		return err
	}

	access := Access {
		OwnerEmail: ownerEmail,
		Viewers: viewers,
		Modifiers: modifiers,
		Certifiers: certifiers,
	}
	user, _, _ := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if org == "Org2MSP" {
		modifiers[user] = currentTime;
	} else if org == "Org3MSP" {
		certifiers[user] = currentTime;
	}

	accessJSON, err := json.Marshal(access)
	if err != nil {
		return err
	}

	ctx.GetStub().PutState("Access_"+chassisNo, accessJSON)

	return ctx.GetStub().PutState("Classic_"+chassisNo, classicJSON)
}

/*
* @param Transaction Context
* @param {String} make The make of the classic
* @param {String} model The model of the classic
* @param {Integer} year The year of manufacturing the classic
* @param {String} licencePlate The license plate of manufacturing the classic
* @param {String} country The registration country of the classic
* @param {String} engineNo The engine number of the classic
* @description UpdateClassic updates the allowed details of a classic. Only owner can use this function
* @returns The classic stored in the world state
*/
func (s *SmartContract) UpdateClassic(ctx contractapi.TransactionContextInterface, chassisNo string, make string, model string, year int, licencePlate string, country string, engineNo string) error {
	classic, err := s.ReadClassic(ctx, chassisNo)
	if err != nil {
		return err
	}
	classic.Make = make
	classic.Model = model
	classic.Year = year
	classic.LicencePlate = licencePlate
	classic.Country = country
	classic.EngineNo = engineNo

	// overwriting original asset with new asset
	classicJSON, err := json.Marshal(classic)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState("Classic_"+chassisNo, classicJSON)
}



/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} newEmail The email of the new owner
* @description Changes the email of the classic's owner (works as ownership change);
* Only the user with the userName equal to the one registered in the classic can perform this action
* @returns The classic stored in the world state with the new email address
*/
func (s *SmartContract) UpdateClassicEmail(ctx contractapi.TransactionContextInterface, chassisNo string, newEmail string) (string, error){
	classic, err := s.ReadClassic(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	oldEmail := classic.OwnerEmail
	classic.OwnerEmail = newEmail

	//Save the prior owner
	oldOwners := classic.PriorOwners
	classic.PriorOwners = append(oldOwners, oldEmail)

	access, err := s.GetAccess(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	access.OwnerEmail = newEmail

	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}
	newJSONAccess, err := json.Marshal(access)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Classic_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}
	err = ctx.GetStub().PutState("Access_"+chassisNo, newJSONAccess)
	if err != nil {
		return "", err
	}

	return string("The owner of the classic "+ chassisNo + " was updated sucessfully."), nil
}


/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} id The id of the restoration procedure
* @param {String} title The title of the restoration procedure
* @param {String} description The description of the restoration procedure
* @param {String[]} photosIds Array of associated photos to the restoration procedure
* @description Creates a new RestorationStep, and adds it to RestorationSteps slice
* @returns The classic stored in the world state with a new restoration step added
*/
func (s *SmartContract) CreateRestorationStep(ctx contractapi.TransactionContextInterface, chassisNo string, title string, description string, photosIds []string, when string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}

	oldSteps := classic.Restorations
	nSteps := len(oldSteps)
	madeBy, ok, error := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if error != nil {
		return "", error
	}
	if !ok {
		return "", error
	}

	stepID := classic.ChassisNo + "_step_"+strconv.Itoa(nSteps)
	step := RestorationStep{
		ID: stepID,
		Title: title,		
		Description: description,
		PhotosIds: photosIds,
		MadeBy: madeBy, 
		When: when,
	}
	
	classic.Restorations = append(oldSteps, step)

	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Classic_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}

	return string(stepID), nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} stepId The id of the restoration procedure
* @description GetStep returns the step with given stepId of the classic stored in the world state with given chassisNo.
* @returns The restoration procedure with the given stepId
*/
func (s *SmartContract) GetStep(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string) (*RestorationStep, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return nil, err
	}
	steps := classic.Restorations
	var step RestorationStep
	found := false
	for i := range steps {
		if steps[i].ID == stepId {
			found = true
			step = steps[i]
			break
		} 
	}
	if found == false {
		return nil, fmt.Errorf("404nostep")
	}
	return &step, nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} stepId The id of the restoration procedure
* @param {String} newTitle The new title to update the restoration procedure
* @param {String} newDescription The new description to update the restoration procedure
* @description UpdateStep updates the title and description of the step with given stepId of the classic stored in the world state with given chassisNo.
* @returns Success message if the step was updated or error otherwise
*/
func (s *SmartContract) UpdateStep(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string, newTitle string, newDescription string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	steps := classic.Restorations
	found := false
	for i := range steps {
		if steps[i].ID == stepId {
			found = true
			user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
			if err != nil {
				return "", err
			}
			if !ok {
				return "", err
			}
			if user == steps[i].MadeBy {
				steps[i].Title = newTitle
				steps[i].Description = newDescription
				break
			} else {
				return "", fmt.Errorf("404madeby")
			}
		} 
	}
	if found == false {
		return "", fmt.Errorf("404nostep")
	}
	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Classic_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}
	return string("The step"+ stepId + " was updated sucessfully."), nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} stepId The id of the restoration procedure
* @param {String[]} photosIds Array of new photos to update the restoration procedure
* @description UpdateStepPhotos updates the photos/documents of the step with given stepId of the classic stored in the world state with given chassisNo.
* @returns Success message if the step was updated or error otherwise
*/
func (s *SmartContract) UpdateStepPhotos(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string, newPhotosIds []string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	steps := classic.Restorations
	found := false
	for i := range steps {
		if steps[i].ID == stepId {
			found = true
			user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
			if err != nil {
				return "", err
			}
			if !ok {
				return "", err
			}
			if user == steps[i].MadeBy {
				oldPhotos := steps[i].PhotosIds
				steps[i].PhotosIds = append(oldPhotos, newPhotosIds...)
				break
			} else {
				return "", fmt.Errorf("404madeby")
			}
		}
	}
	if found == false {
		return "", fmt.Errorf("404nostep")
	}
	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Classic_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}
	return string("The step"+ stepId + " was updated sucessfully."), nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} stepId The id of the restoration procedure
* @param {String} newTitle The new title to update the restoration procedure
* @param {String} newDescription The new description to update the restoration procedure
* @param {String[]} photosIds Array of new photos to update the restoration procedure
* @description UpdateStep updates the title, description and photos of the step with given stepId of the classic stored in the world state with given chassisNo.
* @returns Success message if the step was updated or error otherwise
*/
func (s *SmartContract) UpdateStepAndPhotos(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string, newTitle string, newDescription string, newPhotosIds []string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	steps := classic.Restorations
	found := false
	for i := range steps {
		if steps[i].ID == stepId {
			found = true
			user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
			if err != nil {
				return "", err
			}
			if !ok {
				return "", err
			}
			if user == steps[i].MadeBy {
				steps[i].Title = newTitle
				steps[i].Description = newDescription
				if len(newPhotosIds) > 0 {
					oldPhotos := steps[i].PhotosIds
					steps[i].PhotosIds = append(oldPhotos, newPhotosIds...)
				} 
				break
			} else {
				return "", fmt.Errorf("404madeby")
			}
		} 
	}
	if found == false {
		return "", fmt.Errorf("404nostep")
	}
	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Classic_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}
	return string("The step"+ stepId + " was updated sucessfully."), nil
}
	
/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description DeleteClassic deletes an given classic from the world state.
* DEVELOPING PURPOSES ONLY
*/
func (s *SmartContract) DeleteClassic(ctx contractapi.TransactionContextInterface, chassisNo string) error {
	exists, err := s.ClassicExists(ctx, chassisNo)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("404")
	}

	ctx.GetStub().DelState("Access_"+chassisNo)
	return ctx.GetStub().DelState("Classic_"+chassisNo)
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description Checks if a classic with the chassisNo exists recorded in the ledger
* @returns True if the classic exists, false otherwise
*/
func (s *SmartContract) ClassicExists(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error) {
	classicJSON, err := ctx.GetStub().GetState("Classic_"+chassisNo)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return classicJSON != nil, nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description ReadAsset returns the asset stored in the world state with given chassisNo. Only the owner of the classic has access to this function
* @returns The asset stored in the world state
*/
func (s *SmartContract) ReadClassic(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
	classicJSON, err := ctx.GetStub().GetState("Classic_"+chassisNo)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if classicJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var classic Classic
	err = json.Unmarshal(classicJSON, &classic)
	if err != nil {
		return nil, err
	}

	//DEVELOPING....
	//METER TAMBEM A HIPOTESE DE SER UM ADMIN OU ASSIM
	//hf.Affiliation este attr retorna: org.affiliation
	err = ctx.GetClientIdentity().AssertAttributeValue("hf.EnrollmentID", classic.OwnerEmail)
	if  err != nil {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} documentName The name of the document to be added
* @param {String} documentFile The IPFS CID of the document to be added
* @description Adds a new document to the classic with the given chassisNo
* @returns Success or error
*/
func (s *SmartContract) AddDocument(ctx contractapi.TransactionContextInterface, chassisNo string, documentName string, documentFile string) (string, error){
	classic, err := s.ReadClassicAsDocumenter(ctx, chassisNo)
	if err != nil {
		return "", err
	}

	documents := classic.Documents
	documents[documentName] = documentFile

	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Classic_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}

	return string("The new document "+ documentName + " was added sucessfully."), nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description Function to retrieve the history of a specific classic with the chassisNo as the key
* @returns The history of asset stored in the world state
* USED
*/
func (s *SmartContract) GetClassicHistory2(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (string, error) {
	_, err := s.ReadClassicAsViewer(ctx, chassisNo, currentTime)
	if err != nil {
		return "", err
	}


	historyIterator, err := ctx.GetStub().GetHistoryForKey("Classic_"+chassisNo)
	if err != nil {
		return "", fmt.Errorf("Error in getting the history classic with chassis number %s does not exist", chassisNo)
	}

	counter := 0
	resultJSON := "["
	for historyIterator.HasNext() {
		result, err := historyIterator.Next()
		if err != nil {
			return "", err
		}
		data :="{\"txn\":\"" + result.GetTxId() +"\""
		data +=" ,  \"timestamp\":\"" + string(result.GetTimestamp().AsTime().Format(time.RFC3339))+"\""
		data +=" , \"value\": "+ string(result.GetValue()) + "} "
		if counter > 0 {
			data = ", "+data
		}
		resultJSON += data

		counter++
	}
	historyIterator.Close()

	resultJSON += "]"
	resultJSON = "{ \"counter\": " + strconv.Itoa(counter) + ", \"txns\":" + resultJSON  + "}"
	return resultJSON, nil 
	
}

/*
* @param Transaction Context
* @param {String} owner The email of the owner 
* @description Function to retrieve the all the cars of the user with the specific email
* @returns The list of the cars owned by a specific user
*/
func (s *SmartContract) QueryClassicsByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*Classic, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("Classic_", "Classic_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var classics []*Classic
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var classic Classic
		err = json.Unmarshal(queryResponse.Value, &classic)
		if err != nil {
			return nil, err
		}
		if classic.OwnerEmail == owner {
			classics = append(classics, &classic)
		}
	}
	if len(classics) == 0  {
		return nil, fmt.Errorf("404")
	}
	return classics, nil
}

/*
* @param Transaction Context
* @param {String} owner The email of the workshop/modifier 
* @description Function to retrieve the all the cars of that a workshop/modifier email has access to
* @returns The list of the cars that a workshop/modifier email has access to
*/
func (s *SmartContract) QueryClassicsByModifier(ctx contractapi.TransactionContextInterface, owner string) ([]*Classic, error) {
	modifier, ok, error := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if error != nil {
		return nil, error
	}
	if !ok {
		return nil, error
	}
	resultsIterator, err := ctx.GetStub().GetStateByRange("Access_", "Access_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var classics []*Classic
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var access Access 
		err = json.Unmarshal(queryResponse.Value, &access)
		if err != nil {
			return nil, err
		}
		modifiers := access.Modifiers
		_, ok := modifiers[modifier]
		if ok {
			var classic Classic
			classicKey := strings.Replace(queryResponse.Key, "Access", "Classic", 1)
			classicJSON, err := ctx.GetStub().GetState(classicKey)
			err = json.Unmarshal(classicJSON, &classic)
			if err != nil {
				return nil, err
			}
			classics = append(classics, &classic)
		}
	}
	if len(classics) == 0  {
		return nil, fmt.Errorf("404")
	}
	return classics, nil
}

/*
* @param Transaction Context
* @param {String} owner The email of the certifier entity 
* @description Function to retrieve the all the cars of that a certifier entity email has access to
* @returns The list of the cars that a certifier entity email has access to
*/
func (s *SmartContract) QueryClassicsByCertifier(ctx contractapi.TransactionContextInterface, owner string) ([]*Classic, error) {
	certifier, ok, error := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if error != nil {
		return nil, error
	}
	if !ok {
		return nil, error
	}
	resultsIterator, err := ctx.GetStub().GetStateByRange("Access_", "Access_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var classics []*Classic
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var access Access 
		err = json.Unmarshal(queryResponse.Value, &access)
		if err != nil {
			return nil, err
		}
		certifiers := access.Certifiers
		_, ok := certifiers[certifier]
		if ok {
			var classic Classic
			classicKey := strings.Replace(queryResponse.Key, "Access", "Classic", 1)
			classicJSON, err := ctx.GetStub().GetState(classicKey)
			err = json.Unmarshal(classicJSON, &classic)
			if err != nil {
				return nil, err
			}
			classics = append(classics, &classic)
		}
	}
	if len(classics) == 0  {
		return nil, fmt.Errorf("404")
	}
	return classics, nil
}

//=====ACCESS METHODS======

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description Function to retrieve access permissions of the classic
* @returns The access permissions of the classic
*/
func (s *SmartContract) GetAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (*Access, error) {
	accessJSON, err := ctx.GetStub().GetState("Access_"+chassisNo)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}

	if accessJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return nil, err
	}

	err = ctx.GetClientIdentity().AssertAttributeValue("hf.EnrollmentID", access.OwnerEmail)
	if  err != nil {
		return nil, fmt.Errorf("403")
	}

	return &access, nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} email The email of the user to give access permissions to
* @param {String} level The level of permissions to give. Can be "modifier", "viewer" or "certifier"
* @param {String} currentTime The current time and date as string
* @description Function to give access permissions of the classic to another user
* @returns Success or error
*/
func (s *SmartContract) GiveAccess(ctx contractapi.TransactionContextInterface, chassisNo string, email string, level string, currentTime string) (string, error){
	access, err := s.GetAccess(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	switch level {
	case "viewer":
		viewers := access.Viewers
		viewers[email] = currentTime
	case "modifier":
		modifiers := access.Modifiers
		modifiers[email] = currentTime
	case "certifier":
		certifiers := access.Certifiers
		certifiers[email] = currentTime
	default:
		return "", fmt.Errorf("404nolevel")
	}
	newJSON, err := json.Marshal(access)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Access_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}

	return string("Classic's access permissions were updated!"), nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} email The email of the user to revoke access permissions
* @param {String} level The level of permissions to revoke. Can be "modifier", "viewer" or "certifier"
* @param {String} currentTime The current time and date as string
* @description Function to revoke access permissions of the classic to another user
* @returns Success or error
*/
func (s *SmartContract) RevokeAccess(ctx contractapi.TransactionContextInterface, chassisNo string, email string, level string) (string, error){
	access, err := s.GetAccess(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	switch level {
	case "viewer":
		viewers := access.Viewers
		delete(viewers, email)
	case "modifier":
		modifiers := access.Modifiers
		delete(modifiers, email)
	case "certifier":
		certifiers := access.Certifiers
		delete(certifiers, email)
	default:
		return "", fmt.Errorf("404nolevel")
	}
	newJSON, err := json.Marshal(access)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState("Access_"+chassisNo, newJSON)
	if err != nil {
		return "", err
	}

	return string("Classic's access permissions were revoked!"), nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} currentTime The current time and date as string
* @description Function to check if the user making the request has viewer permissions
* @returns Success if true or false, otherwise
*/
func (s *SmartContract) HasViewerAccess(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (bool, error){
	accessJSON, err := ctx.GetStub().GetState("Access_"+chassisNo)
	if err != nil {
		return false, fmt.Errorf("Failed to read from world state: %v", err)
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return false, err
	}
	if !ok {
		return false, err
	}
	if user == access.OwnerEmail {
		//Owner
		return true, nil
	}

	//Modifiers
	modifiers := access.Modifiers
	_, ok1 := modifiers[user]
	if ok1 {
		return true, nil
	}
	//Certifiers
	certifiers := access.Certifiers
	_, ok2 := certifiers[user]
	if ok2 {
		return true, nil
	}
	//Viewers
	viewers := access.Viewers
	date, ok := viewers[user]
	parsedTime, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return false, err
	}
	if ok {
		parsedTime = parsedTime.Add(1440 * time.Minute)
		t, _ := time.Parse(time.RFC3339, currentTime)
		if t.Before(parsedTime) {
			return true, nil
		}
	}

	return false,  fmt.Errorf("403")
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} currentTime The current time and date as string
* @description ReadClassicAsViewer returns the asset stored in the world state with given chassisNo. Only if the user has viewer access level
* @returns The asset stored in the world state
*/
func (s *SmartContract) ReadClassicAsViewer(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (*Classic, error) {
	classicJSON, err := ctx.GetStub().GetState("Classic_"+chassisNo)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if classicJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var classic Classic
	err = json.Unmarshal(classicJSON, &classic)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.HasViewerAccess(ctx, chassisNo, currentTime)
	if  !hasAccess {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} currentTime The current time and date as string
* @description Function to check if the user making the request has modifier permissions
* @returns Success if true or false, otherwise
*/
func (s *SmartContract) HasModifierAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error){
	accessJSON, err := ctx.GetStub().GetState("Access_"+chassisNo)
	if err != nil {
		return false, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return false, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return false, err
	}
	if !ok {
		return false, err
	}
	if user == access.OwnerEmail {
		//Owner
		return true, nil
	}

	//Modifiers
	modifiers := access.Modifiers
	_, ok1 := modifiers[user]
	if ok1 {
		return true, nil
	}

	return false,  fmt.Errorf("403")
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} currentTime The current time and date as string
* @description ReadClassicAsModifier returns the asset stored in the world state with given chassisNo. Only if the user has modifier access level
* @returns The asset stored in the world state
*/
func (s *SmartContract) ReadClassicAsModifier(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
	classicJSON, err := ctx.GetStub().GetState("Classic_"+chassisNo)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if classicJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var classic Classic
	err = json.Unmarshal(classicJSON, &classic)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.HasModifierAccess(ctx, chassisNo)
	if  !hasAccess {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @param {String} currentTime The current time and date as string
* @description Function to check if the user making the request has certifier permissions
* @returns Success if true or false, otherwise
*/
func (s *SmartContract) HasCertifierAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error){
	accessJSON, err := ctx.GetStub().GetState("Access_"+chassisNo)
	if err != nil {
		return false, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return false, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return false, err
	}
	if !ok {
		return false, err
	}
	//Certifiers only
	certifiers := access.Certifiers
	_, ok1 := certifiers[user]
	if ok1 {
		return true, nil
	}

	return false,  fmt.Errorf("403")
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description Marks the classic as certified by the user making the request. Only available for users with certifier access level
* @returns The asset stored in the world state
*/
func (s *SmartContract) MarkAsCertified(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
	classicJSON, err := ctx.GetStub().GetState("Classic_"+chassisNo)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if classicJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var classic Classic
	err = json.Unmarshal(classicJSON, &classic)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.HasCertifierAccess(ctx, chassisNo)
	if  !hasAccess {
		return nil, fmt.Errorf("403")
	} else {
		user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, err
		}
		oldCertifications := classic.Certifications
		classic.Certifications = append(oldCertifications, user)

		newJSON, err := json.Marshal(classic)
		if err != nil {
			return nil, err
		}
		err = ctx.GetStub().PutState("Classic_"+chassisNo, newJSON)
		if err != nil {
			return nil, err
		}
		return &classic, nil
	}	
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description Function to check if the user making the request has documenter permissions
* @returns Success if true or false, otherwise
*/
func (s *SmartContract) HasDocumenterAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error){
	accessJSON, err := ctx.GetStub().GetState("Access_"+chassisNo)
	if err != nil {
		return false, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return false, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return false, err
	}
	if !ok {
		return false, err
	}
	if user == access.OwnerEmail {
		//Owner
		return true, nil
	}

	//Modifiers
	modifiers := access.Modifiers
	_, ok1 := modifiers[user]
	if ok1 {
		return true, nil
	}

	certifiers := access.Certifiers
	_, ok2 := certifiers[user]
	if ok2 {
		return true, nil
	}

	return false,  fmt.Errorf("403")
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description ReadClassicAsDocumenter returns the asset stored in the world state with given chassisNo. Only if the user has documenter access level
* @returns The asset stored in the world state
*/
func (s *SmartContract) ReadClassicAsDocumenter(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
	classicJSON, err := ctx.GetStub().GetState("Classic_"+chassisNo)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if classicJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var classic Classic
	err = json.Unmarshal(classicJSON, &classic)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.HasDocumenterAccess(ctx, chassisNo)
	if  !hasAccess {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description Function to check the level of access the requesting user has regarding the classic with chassisNo.
* @returns String representing the access level
*/
func (s *SmartContract) CheckUserAccess(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (string, error) {
	accessJSON, err := ctx.GetStub().GetState("Access_"+chassisNo)
	if err != nil {
		return "", fmt.Errorf("Failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return "", fmt.Errorf("404")
	}
	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return "", err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", err
	}
	if user == access.OwnerEmail {
		//Owner
		return string("owner"), nil
	}

	//Modifiers
	modifiers := access.Modifiers
	_, ok1 := modifiers[user]
	if ok1 {
		return string("modifier"), nil
	}
	//Certifiers
	certifiers := access.Certifiers
	_, ok2 := certifiers[user]
	if ok2 {
		return string("certifier"), nil
	}
	//Viewers
	viewers := access.Viewers
	date, ok := viewers[user]
	parsedTime, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return "", err
	}
	if ok {
		parsedTime = parsedTime.Add(1440 * time.Minute)
		t, _ := time.Parse(time.RFC3339, currentTime)
		if t.Before(parsedTime) {
			return string("viewer"), nil
		}
	}
	return "", fmt.Errorf("403")
}

/*
* @param Transaction Context
* @param {String} chassisNo The chassis number of a classic
* @description Function to retrieve the history of the access permissions. Function only available to the onwer.
* @returns The access' history of the classic
*/
func (s *SmartContract) GetAccessHistory(ctx contractapi.TransactionContextInterface, chassisNo string) (string, error) {
	_, err := s.GetAccess(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	historyIterator, err := ctx.GetStub().GetHistoryForKey("Access_"+chassisNo)
	if err != nil {
		return "", fmt.Errorf("Error in getting the history of access permissions of classic with chassis number %s does not exist", chassisNo)
	}

	counter := 0
	resultJSON := "["
	for historyIterator.HasNext() {
		result, err := historyIterator.Next()
		if err != nil {
			return "", err
		}
		data :="{\"txn\":\"" + result.GetTxId() +"\""
		data +=" ,  \"timestamp\":\"" + string(result.GetTimestamp().AsTime().Format(time.RFC3339))+"\""
		data +=" , \"value\": "+ string(result.GetValue()) + "} "
		if counter > 0 {
			data = ", "+data
		}
		resultJSON += data

		counter++
	}

	historyIterator.Close()
	resultJSON += "]"
	resultJSON = "{ \"counter\": " + strconv.Itoa(counter) + ", \"txns\":" + resultJSON  + "}"
	return resultJSON, nil 	
}


/*
Function to print info about the creater of transaction
*/
func PrintCreatorInfo(ctx contractapi.TransactionContextInterface) {
	fmt.Println("PrintCreatorInfo() executed ")
	byteData, _ := ctx.GetStub().GetCreator()
	fmt.Println("PrintCreatorInfo => ",string(byteData))
	fmt.Println("-----------------------ENDZAOO-----------------")
}

/*
Auxiliar function to print the ID and MSPID of the proposing client
NOT USED at the moment
*/
func PrintIDInfo(ctx contractapi.TransactionContextInterface) {
	id, _:= cid.GetID(ctx.GetStub())
	mspid, _ := cid.GetMSPID(ctx.GetStub())
	fmt.Println("----------------INicio.-----------")
	fmt.Println( "The id is "+id+"; and the mspid is "+mspid+".")
	fmt.Println("----------------FIMMMMMM.-----------")
}

/*
Function to get the decoded info about the client in this format:
x509::CN=user1org1,OU=org1+OU=client+OU=department1::CN=ca.org1.example.com,O=org1.example.com,L=Durham,ST=North Carolina,C=US
*/
func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodedID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("Failed to base64 decode clientID: %v", err)
	}
	return string(decodedID), nil
}



/*
Auxiliar function to check if the network is still runing
Uses the function GetSubmittingClientIdentity
*/
func (s *SmartContract) HealthCheck(ctx contractapi.TransactionContextInterface) (string, error) {
	fmt.Println("----------------INicio.-----------")
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to get clientID: %v", err)
	}
	fmt.Println( "The id is "+clientID)
	fmt.Println("----------------FIMMMMMM.-----------")
	return "Health checked!", nil
}

/*
=====Main Function=====
*/
func main() {
	// See chaincode.env.example
	config := ServerConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create classiccars chaincode: %s", err.Error())
		return
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		fmt.Printf("Error starting classiccars chaincode: %s", err.Error())
	}
}
