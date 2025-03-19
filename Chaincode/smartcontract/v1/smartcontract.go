package v1

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"

	"encoding/base64"
)

const (
	Org2MSP = "Org2MSP"
	Org3MSP = "Org3MSP"
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
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PhotosIds   []string `json:"photosIds"`
	MadeBy      string   `json:"madeBy"`
	When        string   `json:"when"`
}

// Classic struct describes basic details of what makes up a classic, chassisNo is the main key
type Classic struct {
	Make           string            `json:"make"`
	Model          string            `json:"model"`
	Year           int               `json:"year"`
	LicencePlate   string            `json:"licencePlate"`
	Country        string            `json:"country"`
	ChassisNo      string            `json:"chassisNo"`
	EngineNo       string            `json:"engineNo"`
	OwnerEmail     string            `json:"ownerEmail"`
	PriorOwners    []string          `json:"priorOwners"`
	Restorations   []RestorationStep `json:"restorations"`
	Certifications []string          `json:"certifications"`
	Documents      map[string]string `json:"documents"`
}

// Access struct describes access permission details
type Access struct {
	OwnerEmail string            `json:"ownerEmail"`
	Viewers    map[string]string `json:"viewers"`
	Modifiers  map[string]string `json:"modifiers"`
	Certifiers map[string]string `json:"certifiers"`
}

// InitLedger initiates the ledger, adding a base classic to it
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	priorOwners := []string{}
	restorations := []RestorationStep{}
	certifications := []string{}
	documents := map[string]string{}
	classics := []Classic{
		// TODO check why this is needed
		{Make: "Jaguar", Model: "E", Year: 1964, LicencePlate: "AWA 69B", Country: "Portugal", ChassisNo: "07642", EngineNo: "ER8595", OwnerEmail: "raimundo.branco@yopmail.com", PriorOwners: priorOwners, Restorations: restorations, Certifications: certifications, Documents: documents},
	}

	for _, classic := range classics {
		classicJSON, err := json.Marshal(classic)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(classicKey(classic.ChassisNo), classicJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	PrintCreatorInfo(ctx)

	return nil
}

// PrintCreatorInfo is a debug function that prints the creator of the transaction
func PrintCreatorInfo(ctx contractapi.TransactionContextInterface) {
	byteData, _ := ctx.GetStub().GetCreator()
	fmt.Println("PrintCreatorInfo => ", string(byteData))
}

// GetSubmittingClientIdentity returns decoded client identity in the format:
// x509::CN=user1org1,OU=org1+OU=client+OU=department1::CN=ca.org1.example.com,O=org1.example.com,L=Durham,ST=North Carolina,C=US
func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to read clientID: %v", err)
	}

	decodedID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodedID), nil
}

// HealthCheck function to check the health of the network
func (s *SmartContract) HealthCheck(ctx contractapi.TransactionContextInterface) (string, error) {
	_, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get clientID: %v", err)
	}
	return "Health checked!", nil
}

func accessKey(chassisNo string) string {
	return fmt.Sprintf("Access_%s", chassisNo)
}

func classicKey(chassisNo string) string {
	return fmt.Sprintf("Classic_%s", chassisNo)
}

func (s *SmartContract) getClassic(ctx contractapi.TransactionContextInterface, chassisNo string) (*Classic, error) {
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
		return nil, fmt.Errorf("failed to unmarshal Classic JSON: %v", err)
	}

	return &classic, nil
}
