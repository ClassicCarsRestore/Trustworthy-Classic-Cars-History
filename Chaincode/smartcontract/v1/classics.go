package v1

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"strconv"
	"strings"
	"time"
)

// GetAllClassics returns all classics found in world state
func (s *SmartContract) GetAllClassics(ctx contractapi.TransactionContextInterface) ([]*Classic, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("Classic_", "Classic_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	classics := []*Classic{}
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

// CreateClassic issues a new asset to the world state with given details. Only a verified workshop can perform this function
// TODO refactor to receive struct
func (s *SmartContract) CreateClassic(ctx contractapi.TransactionContextInterface, make string, model string, year int, licencePlate string, country string, chassisNo string, engineNo string, ownerEmail string, currentTime string) error {
	org, ok, err := ctx.GetClientIdentity().GetAttributeValue("org")
	if err != nil {
		return err
	}
	if !ok {
		// TODO needs proper error handling
		return fmt.Errorf("403")
	}
	if org != Org2MSP && org != Org3MSP {
		return fmt.Errorf("403")
	}

	exists, err := s.ClassicExists(ctx, chassisNo)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("409")
	}

	classic := Classic{
		Make:           make,
		Model:          model,
		Year:           year,
		LicencePlate:   licencePlate,
		Country:        country,
		ChassisNo:      chassisNo,
		EngineNo:       engineNo,
		OwnerEmail:     ownerEmail,
		PriorOwners:    []string{},
		Restorations:   []RestorationStep{},
		Certifications: []string{},
		Documents:      map[string]string{},
	}
	classicJSON, err := json.Marshal(classic)
	if err != nil {
		return err
	}

	access := Access{
		OwnerEmail: ownerEmail,
		Viewers:    map[string]string{},
		Modifiers:  map[string]string{},
		Certifiers: map[string]string{},
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return err
	}
	if !ok {
		// TODO needs proper error handling
		return fmt.Errorf("403")
	}
	if org == Org2MSP {
		access.Modifiers[user] = currentTime
	} else if org == Org3MSP {
		access.Certifiers[user] = currentTime
	}

	accessJSON, err := json.Marshal(access)
	if err != nil {
		return err
	}

	// TODO needs error handling
	err = ctx.GetStub().PutState(accessKey(chassisNo), accessJSON)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(classicKey(chassisNo), classicJSON)
}

// UpdateClassic updates the allowed details of a classic. Only owner can use this function
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

	classicJSON, err := json.Marshal(classic)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(classicKey(chassisNo), classicJSON)
}

// UpdateClassicEmail changes the email of the classic's owner
// Only the user with the userName equal to the one registered in the classic can perform this action
func (s *SmartContract) UpdateClassicEmail(ctx contractapi.TransactionContextInterface, chassisNo string, newEmail string) (string, error) {
	classic, err := s.ReadClassic(ctx, chassisNo)
	if err != nil {
		return "", err
	}

	classic.OwnerEmail = newEmail
	classic.PriorOwners = append(classic.PriorOwners, classic.OwnerEmail)

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

	err = ctx.GetStub().PutState(classicKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}
	err = ctx.GetStub().PutState(accessKey(chassisNo), newJSONAccess)
	if err != nil {
		return "", err
	}

	return "The owner of the classic " + chassisNo + " was updated sucessfully.", nil
}

// CreateRestorationStep creates a new restoration step and adds it to the classic stored in the world state with given chassisNo.
func (s *SmartContract) CreateRestorationStep(ctx contractapi.TransactionContextInterface, chassisNo string, title string, description string, photosIds []string, when string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}

	madeBy, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return "", err
	}
	if !ok {
		// TODO needs proper error handling
		return "", err
	}

	stepID := fmt.Sprintf("%s_step_%d", chassisNo, len(classic.Restorations))
	step := RestorationStep{
		ID:          stepID,
		Title:       title,
		Description: description,
		PhotosIds:   photosIds,
		MadeBy:      madeBy,
		When:        when,
	}

	classic.Restorations = append(classic.Restorations, step)

	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(classicKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}

	return stepID, nil
}

// GetStep returns the step with given stepId of the classic stored in the world state with given chassisNo.
func (s *SmartContract) GetStep(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string) (*RestorationStep, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return nil, err
	}

	var step RestorationStep
	found := false
	for i := range classic.Restorations {
		if classic.Restorations[i].ID == stepId {
			found = true
			step = classic.Restorations[i]
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("404nostep")
	}

	return &step, nil
}

// UpdateStep updates the title and description of the step with given stepId of the classic stored in the world state with given chassisNo.
func (s *SmartContract) UpdateStep(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string, newTitle string, newDescription string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}

	found := false
	for i := range classic.Restorations {
		if classic.Restorations[i].ID == stepId {
			found = true
			user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
			if err != nil {
				return "", err
			}
			if !ok {
				// TODO needs proper error handling
				return "", err
			}
			if user == classic.Restorations[i].MadeBy {
				classic.Restorations[i].Title = newTitle
				classic.Restorations[i].Description = newDescription
				break
			} else {
				return "", fmt.Errorf("404madeby")
			}
		}
	}

	if !found {
		return "", fmt.Errorf("404nostep")
	}
	newJSON, err := json.Marshal(classic)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(classicKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}
	return "The step" + stepId + " was updated sucessfully.", nil
}

// UpdateStepPhotos updates the photos of the step with given stepId of the classic stored in the world state with given chassisNo.
func (s *SmartContract) UpdateStepPhotos(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string, newPhotosIds []string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}

	found := false
	for i := range classic.Restorations {
		if classic.Restorations[i].ID == stepId {
			found = true
			user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
			if err != nil {
				return "", err
			}
			if !ok {
				return "", err
			}
			if user == classic.Restorations[i].MadeBy {
				classic.Restorations[i].PhotosIds = append(classic.Restorations[i].PhotosIds, newPhotosIds...)
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

	err = ctx.GetStub().PutState(classicKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}
	return "The step" + stepId + " was updated sucessfully.", nil
}

// UpdateStepAndPhotos updates the title, description and photos of the step with given stepId of the classic stored in the world state with given chassisNo.
// TODO check if needed; redundant with UpdateStep and UpdateStepPhotos
func (s *SmartContract) UpdateStepAndPhotos(ctx contractapi.TransactionContextInterface, chassisNo string, stepId string, newTitle string, newDescription string, newPhotosIds []string) (string, error) {
	classic, err := s.ReadClassicAsModifier(ctx, chassisNo)
	if err != nil {
		return "", err
	}

	found := false
	for i := range classic.Restorations {
		if classic.Restorations[i].ID == stepId {
			found = true
			user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
			if err != nil {
				return "", err
			}
			if !ok {
				return "", err
			}
			if user == classic.Restorations[i].MadeBy {
				classic.Restorations[i].Title = newTitle
				classic.Restorations[i].Description = newDescription
				if len(newPhotosIds) > 0 {
					classic.Restorations[i].PhotosIds = append(classic.Restorations[i].PhotosIds, newPhotosIds...)
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

	err = ctx.GetStub().PutState(classicKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}
	return "The step" + stepId + " was updated sucessfully.", nil
}

// DeleteClassic deletes a given classic from the world state. DEVELOPING PURPOSES ONLY
// TODO check if actually needed
func (s *SmartContract) DeleteClassic(ctx contractapi.TransactionContextInterface, chassisNo string) error {
	exists, err := s.ClassicExists(ctx, chassisNo)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("404")
	}

	err = ctx.GetStub().DelState(accessKey(chassisNo))
	if err != nil {
		return err
	}
	return ctx.GetStub().DelState(classicKey(chassisNo))
}

// ClassicExists returns true when classic with given chassisNo exists in world state
func (s *SmartContract) ClassicExists(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error) {
	classicJSON, err := ctx.GetStub().GetState(classicKey(chassisNo))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return classicJSON != nil, nil
}

// ReadClassic returns the classic with the given chassisNo
func (s *SmartContract) ReadClassic(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
	classicJSON, err := ctx.GetStub().GetState(classicKey(chassisNo))
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if classicJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var classic Classic
	err = json.Unmarshal(classicJSON, &classic)
	if err != nil {
		return nil, err
	}

	// TODO support action by admin
	err = ctx.GetClientIdentity().AssertAttributeValue(enrollmentIDAtt, classic.OwnerEmail)
	if err != nil {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

// AddDocument adds a new document to the classic with the given chassisNo
func (s *SmartContract) AddDocument(ctx contractapi.TransactionContextInterface, chassisNo string, documentName string, documentFile string) (string, error) {
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

	err = ctx.GetStub().PutState(classicKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}

	return "The new document " + documentName + " was added sucessfully.", nil
}

// GetClassicHistory2 returns the history of a classic with the given chassisNo
func (s *SmartContract) GetClassicHistory2(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (string, error) {
	_, err := s.ReadClassicAsViewer(ctx, chassisNo, currentTime)
	if err != nil {
		return "", err
	}

	historyIterator, err := ctx.GetStub().GetHistoryForKey(classicKey(chassisNo))
	if err != nil {
		return "", fmt.Errorf("error in getting the history classic with chassis number %s does not exist", chassisNo)
	}

	counter := 0
	resultJSON := "["
	for historyIterator.HasNext() {
		result, err := historyIterator.Next()
		if err != nil {
			return "", err
		}
		data := "{\"txn\":\"" + result.GetTxId() + "\""
		data += " ,  \"timestamp\":\"" + string(result.GetTimestamp().AsTime().Format(time.RFC3339)) + "\""
		data += " , \"value\": " + string(result.GetValue()) + "} "
		if counter > 0 {
			data = ", " + data
		}
		resultJSON += data

		counter++
	}
	historyIterator.Close()

	resultJSON += "]"
	resultJSON = "{ \"counter\": " + strconv.Itoa(counter) + ", \"txns\":" + resultJSON + "}"
	return resultJSON, nil

}

// QueryClassicsByOwner returns all classics that a user owns
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
	if len(classics) == 0 {
		return nil, fmt.Errorf("404")
	}
	return classics, nil
}

// QueryClassicsByModifier returns all classics that a modifier entity has access to
func (s *SmartContract) QueryClassicsByModifier(ctx contractapi.TransactionContextInterface, owner string) ([]*Classic, error) {
	modifier, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("403")
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
	if len(classics) == 0 {
		return nil, fmt.Errorf("404")
	}
	return classics, nil
}

// QueryClassicsByCertifier returns all classics that a certifier entity has access to
// TODO check if safe to remove unused parameter
func (s *SmartContract) QueryClassicsByCertifier(ctx contractapi.TransactionContextInterface, owner string) ([]*Classic, error) {
	certifier, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("403")
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
			key := strings.Replace(queryResponse.Key, "Access", "Classic", 1)
			classicJSON, err := ctx.GetStub().GetState(key)
			err = json.Unmarshal(classicJSON, &classic)
			if err != nil {
				return nil, err
			}
			classics = append(classics, &classic)
		}
	}
	if len(classics) == 0 {
		return nil, fmt.Errorf("404")
	}
	return classics, nil
}
