package v1

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"strconv"
	"time"
)

const (
	enrollmentIDAtt      = "hf.EnrollmentID"
	accessLevelOwner     = "owner"
	accessLevelViewer    = "viewer"
	accessLevelModifier  = "modifier"
	accessLevelCertifier = "certifier"
)

// GetAccess returns the access permissions of the classic
func (s *SmartContract) GetAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (*Access, error) {
	accessJSON, err := ctx.GetStub().GetState(accessKey(chassisNo))
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if accessJSON == nil {
		return nil, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return nil, err
	}

	err = ctx.GetClientIdentity().AssertAttributeValue(enrollmentIDAtt, access.OwnerEmail)
	if err != nil {
		return nil, fmt.Errorf("403")
	}

	return &access, nil
}

// GiveAccess gives access permissions of the classic to another user
func (s *SmartContract) GiveAccess(ctx contractapi.TransactionContextInterface, chassisNo string, email string, level string, currentTime string) (string, error) {
	access, err := s.GetAccess(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	switch level {
	case accessLevelViewer:
		viewers := access.Viewers
		viewers[email] = currentTime
	case accessLevelModifier:
		modifiers := access.Modifiers
		modifiers[email] = currentTime
	case accessLevelCertifier:
		certifiers := access.Certifiers
		certifiers[email] = currentTime
	default:
		return "", fmt.Errorf("404nolevel")
	}
	newJSON, err := json.Marshal(access)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(accessKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}

	return "Classic's access permissions were updated!", nil
}

// RevokeAccess revokes access permissions of the classic to another user
func (s *SmartContract) RevokeAccess(ctx contractapi.TransactionContextInterface, chassisNo string, email string, level string) (string, error) {
	access, err := s.GetAccess(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	switch level {
	case accessLevelViewer:
		viewers := access.Viewers
		delete(viewers, email)
	case accessLevelModifier:
		modifiers := access.Modifiers
		delete(modifiers, email)
	case accessLevelCertifier:
		certifiers := access.Certifiers
		delete(certifiers, email)
	default:
		return "", fmt.Errorf("404nolevel")
	}

	newJSON, err := json.Marshal(access)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(accessKey(chassisNo), newJSON)
	if err != nil {
		return "", err
	}

	return "Classic's access permissions revoked!", nil
}

// HasViewerAccess checks if the user making the request has viewer access level
func (s *SmartContract) HasViewerAccess(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (bool, error) {
	accessJSON, err := ctx.GetStub().GetState(accessKey(chassisNo))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("403")
	}
	if user == access.OwnerEmail {
		return true, nil
	}

	_, ok = access.Modifiers[user]
	if ok {
		return true, nil
	}

	_, ok = access.Certifiers[user]
	if ok {
		return true, nil
	}

	date, ok := access.Viewers[user]
	if ok {
		parsedTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return false, err
		}

		parsedTime = parsedTime.Add(1440 * time.Minute)
		t, _ := time.Parse(time.RFC3339, currentTime)
		if t.Before(parsedTime) {
			return true, nil
		}
	}

	return false, fmt.Errorf("403")
}

// ReadClassicAsViewer returns the asset stored in the world state with given chassisNo. Only if the user has viewer access level
func (s *SmartContract) ReadClassicAsViewer(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (*Classic, error) {
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

	hasAccess, err := s.HasViewerAccess(ctx, chassisNo, currentTime)
	if !hasAccess {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

// HasModifierAccess checks if the user making the request has modifier access level
func (s *SmartContract) HasModifierAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error) {
	accessJSON, err := ctx.GetStub().GetState(accessKey(chassisNo))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return false, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("403")
	}
	if user == access.OwnerEmail {
		return true, nil
	}

	_, ok = access.Modifiers[user]
	if ok {
		return true, nil
	}

	return false, fmt.Errorf("403")
}

// ReadClassicAsModifier returns the asset stored in the world state with given chassisNo. Only if the user has modifier access level
func (s *SmartContract) ReadClassicAsModifier(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
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

	hasAccess, err := s.HasModifierAccess(ctx, chassisNo)
	if !hasAccess {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

// HasCertifierAccess checks if the user making the request has certifier access level
func (s *SmartContract) HasCertifierAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error) {
	accessJSON, err := ctx.GetStub().GetState(accessKey(chassisNo))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return false, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("403")
	}

	_, ok = access.Certifiers[user]
	if ok {
		return true, nil
	}

	return false, fmt.Errorf("403")
}

// MarkAsCertified marks the classic as certified by the user making the request
func (s *SmartContract) MarkAsCertified(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
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

	hasAccess, err := s.HasCertifierAccess(ctx, chassisNo)
	if !hasAccess {
		return nil, fmt.Errorf("403")
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("403")
	}
	classic.Certifications = append(classic.Certifications, user)

	newJSON, err := json.Marshal(classic)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(classicKey(chassisNo), newJSON)
	if err != nil {
		return nil, err
	}
	return &classic, nil
}

// HasDocumenterAccess checks if the user making the request has documenter access level
func (s *SmartContract) HasDocumenterAccess(ctx contractapi.TransactionContextInterface, chassisNo string) (bool, error) {
	accessJSON, err := ctx.GetStub().GetState(accessKey(chassisNo))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return false, fmt.Errorf("404")
	}

	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return false, err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("403")
	}
	if user == access.OwnerEmail {
		return true, nil
	}

	_, ok = access.Modifiers[user]
	if ok {
		return true, nil
	}

	_, ok = access.Certifiers[user]
	if ok {
		return true, nil
	}

	return false, fmt.Errorf("403")
}

// ReadClassicAsDocumenter returns the asset stored in the world state with given chassisNo. Only if the user has documenter access level
func (s *SmartContract) ReadClassicAsDocumenter(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
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

	hasAccess, err := s.HasDocumenterAccess(ctx, chassisNo)
	if !hasAccess {
		return nil, fmt.Errorf("403")
	}

	return &classic, nil
}

// CheckUserAccess checks the level of access the requesting user has regarding the classic with chassisNo
func (s *SmartContract) CheckUserAccess(ctx contractapi.TransactionContextInterface, chassisNo string, currentTime string) (string, error) {
	accessJSON, err := ctx.GetStub().GetState(accessKey(chassisNo))
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	if accessJSON == nil {
		return "", fmt.Errorf("404")
	}
	var access Access
	err = json.Unmarshal(accessJSON, &access)
	if err != nil {
		return "", err
	}

	user, ok, err := ctx.GetClientIdentity().GetAttributeValue(enrollmentIDAtt)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("403")
	}
	if user == access.OwnerEmail {
		return accessLevelOwner, nil
	}

	_, ok = access.Modifiers[user]
	if ok {
		return accessLevelModifier, nil
	}

	_, ok = access.Certifiers[user]
	if ok {
		return accessLevelCertifier, nil
	}

	date, ok := access.Viewers[user]
	if ok {
		parsedTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return "", err
		}

		parsedTime = parsedTime.Add(1440 * time.Minute)
		t, _ := time.Parse(time.RFC3339, currentTime)
		if t.Before(parsedTime) {
			return accessLevelViewer, nil
		}
	}

	return "", fmt.Errorf("403")
}

// GetAccessHistory returns the history of access permissions of the classic with the given chassisNo
// Only the owner of the classic has access to this function
func (s *SmartContract) GetAccessHistory(ctx contractapi.TransactionContextInterface, chassisNo string) (string, error) {
	_, err := s.GetAccess(ctx, chassisNo)
	if err != nil {
		return "", err
	}
	historyIterator, err := ctx.GetStub().GetHistoryForKey(accessKey(chassisNo))
	if err != nil {
		return "", fmt.Errorf("failed to get access permissions history of classic %s", chassisNo)
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
