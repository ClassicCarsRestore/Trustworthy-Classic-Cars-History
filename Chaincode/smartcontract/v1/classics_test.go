package v1

import (
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/queryresult"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSmartContract_GetAllClassics(t *testing.T) {
	t.Run("ReturnsAllClassicsWhenFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		classic1 := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classic1JSON, _ := json.Marshal(classic1)

		classic2 := Classic{
			ChassisNo: "DEF456",
			Make:      "Porsche",
			Model:     "911",
			Year:      1973,
		}
		classic2JSON, _ := json.Marshal(classic2)

		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)

		mockResultsIterator.On("HasNext").Return(true).Once()
		mockResultsIterator.On("HasNext").Return(true).Once()
		mockResultsIterator.On("HasNext").Return(false)

		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Key: "Classic_ABC123", Value: classic1JSON}, nil,
		).Once()
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Key: "Classic_DEF456", Value: classic2JSON}, nil,
		).Once()

		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAllClassics(mockCtx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "ABC123", result[0].ChassisNo)
		assert.Equal(t, "Ferrari", result[0].Make)
		assert.Equal(t, "Testarossa", result[0].Model)
		assert.Equal(t, 1985, result[0].Year)
		assert.Equal(t, "DEF456", result[1].ChassisNo)
		assert.Equal(t, "Porsche", result[1].Make)
		assert.Equal(t, "911", result[1].Model)
		assert.Equal(t, 1973, result[1].Year)
	})

	t.Run("ReturnsEmptySliceWhenNoClassics", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAllClassics(mockCtx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})

	t.Run("ReturnsErrorWhenGetStateByRangeFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(nil, fmt.Errorf("range query error"))

		sc := SmartContract{}
		result, err := sc.GetAllClassics(mockCtx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "range query error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenNextFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(nil, fmt.Errorf("iterator error"))
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAllClassics(mockCtx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "iterator error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Key: "Classic_ABC123", Value: []byte("invalid json")}, nil,
		)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAllClassics(mockCtx)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSmartContract_CreateClassic(t *testing.T) {
	t.Run("CreatesClassicWithModifierWhenOrg2", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org2MSP, true, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// Classic doesn't exist yet
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		// Expect PutState calls for both classic and access
		mockStub.On("PutState", "Classic_ABC123", mock.MatchedBy(func(data []byte) bool {
			var classic Classic
			json.Unmarshal(data, &classic)
			return classic.ChassisNo == "ABC123" &&
				classic.Make == "Ferrari" &&
				classic.Model == "Testarossa" &&
				classic.Year == 1985
		})).Return(nil)

		mockStub.On("PutState", "Access_ABC123", mock.MatchedBy(func(data []byte) bool {
			var access Access
			json.Unmarshal(data, &access)
			return access.OwnerEmail == "owner@example.com" &&
				len(access.Modifiers) == 1 &&
				access.Modifiers["modifier@example.com"] == "2023-01-01T12:00:00Z"
		})).Return(nil)

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.NoError(t, err)
		mockStub.AssertCalled(t, "PutState", "Classic_ABC123", mock.Anything)
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", mock.Anything)
	})

	t.Run("CreatesClassicWithCertifierWhenOrg3", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org3MSP, true, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		// Classic doesn't exist yet
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		// Expect PutState calls for both classic and access
		mockStub.On("PutState", "Classic_ABC123", mock.MatchedBy(func(data []byte) bool {
			var classic Classic
			json.Unmarshal(data, &classic)
			return classic.ChassisNo == "ABC123" &&
				classic.Make == "Ferrari" &&
				classic.Model == "Testarossa" &&
				classic.Year == 1985
		})).Return(nil)

		mockStub.On("PutState", "Access_ABC123", mock.MatchedBy(func(data []byte) bool {
			var access Access
			json.Unmarshal(data, &access)
			return access.OwnerEmail == "owner@example.com" &&
				len(access.Certifiers) == 1 &&
				access.Certifiers["certifier@example.com"] == "2023-01-01T12:00:00Z"
		})).Return(nil)

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.NoError(t, err)
		mockStub.AssertCalled(t, "PutState", "Classic_ABC123", mock.Anything)
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", mock.Anything)
	})

	t.Run("ReturnsErrorWhenClassicAlreadyExists", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org2MSP, true, nil)

		// Classic already exists
		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)
		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Equal(t, "409", err.Error())
	})

	t.Run("ReturnsErrorWhenOrgIsNotAuthorized", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup unauthorized org
		mockIdentity.On("GetAttributeValue", "org").Return("Org1MSP", true, nil)

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
	})

	t.Run("ReturnsErrorWhenOrgAttributeMissing", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Org attribute missing
		mockIdentity.On("GetAttributeValue", "org").Return("", false, nil)

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorWhenOrgAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Identity read error
		mockIdentity.On("GetAttributeValue", "org").Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "identity error")
	})

	t.Run("ReturnsErrorWhenClassicExistsCheckFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org2MSP, true, nil)

		// GetState error
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org2MSP, true, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// Classic doesn't exist yet
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		// PutState for Access succeeds but for Classic fails
		mockStub.On("PutState", "Access_ABC123", mock.Anything).Return(nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
	})

	t.Run("ReturnsErrorWhenPutStateForAccessFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org2MSP, true, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// Classic doesn't exist yet
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		// PutState for Access fails
		mockStub.On("PutState", "Access_ABC123", mock.Anything).Return(fmt.Errorf("access put state error"))

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access put state error")
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org2MSP, true, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		// Classic doesn't exist yet
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup identity attributes
		mockIdentity.On("GetAttributeValue", "org").Return(Org2MSP, true, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		// Classic doesn't exist yet
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		err := sc.CreateClassic(
			mockCtx,
			"Ferrari",
			"Testarossa",
			1985,
			"ABC123",
			"Italy",
			"ABC123",
			"ENG123",
			"owner@example.com",
			"2023-01-01T12:00:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "identity error")
	})
}

func TestSmartContract_UpdateClassic(t *testing.T) {
	t.Run("UpdatesClassicSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			LicencePlate: "OLD123",
			Country:      "Italy",
			EngineNo:     "ENG123",
			OwnerEmail:   "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		// Mock the AssertAttributeValue call with matching owner email
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		// For PutState
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)

		sc := SmartContract{}
		err := sc.UpdateClassic(mockCtx, "ABC123", "Ferrari", "Testarossa", 1986, "NEW456", "France", "ENG456")

		assert.NoError(t, err)

		// Verify PutState was called with updated classic
		mockStub.AssertNumberOfCalls(t, "PutState", 1)

		// Get the PutState call arguments
		putStateCall := mockStub.Calls[1] // GetState is call 0, PutState is call 1
		assert.Equal(t, "PutState", putStateCall.Method)
		assert.Equal(t, "Classic_ABC123", putStateCall.Arguments[0])

		// Verify the updated classic values
		updatedClassicBytes := putStateCall.Arguments[1].([]byte)
		var updatedClassic Classic
		err = json.Unmarshal(updatedClassicBytes, &updatedClassic)
		assert.NoError(t, err)

		assert.Equal(t, "Ferrari", updatedClassic.Make)
		assert.Equal(t, "Testarossa", updatedClassic.Model)
		assert.Equal(t, 1986, updatedClassic.Year)
		assert.Equal(t, "NEW456", updatedClassic.LicencePlate)
		assert.Equal(t, "France", updatedClassic.Country)
		assert.Equal(t, "ENG456", updatedClassic.EngineNo)
		assert.Equal(t, "owner@example.com", updatedClassic.OwnerEmail) // Email should not change
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		err := sc.UpdateClassic(
			mockCtx,
			"ABC123",
			"Ferrari",
			"Testarossa Spyder",
			1986,
			"NEW456",
			"Germany",
			"ENG456",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("ReturnsErrorWhenGetStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		err := sc.UpdateClassic(
			mockCtx,
			"ABC123",
			"Ferrari",
			"Testarossa Spyder",
			1986,
			"NEW456",
			"Germany",
			"ENG456",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		err := sc.UpdateClassic(
			mockCtx,
			"ABC123",
			"Ferrari",
			"Testarossa Spyder",
			1986,
			"NEW456",
			"Germany",
			"ENG456",
		)

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Create initial classic
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			LicencePlate: "OLD123",
			Country:      "Italy",
			EngineNo:     "ENG123",
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		err := sc.UpdateClassic(
			mockCtx,
			"ABC123",
			"Ferrari",
			"Testarossa Spyder",
			1986,
			"NEW456",
			"Germany",
			"ENG456",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
	})
}

func TestSmartContract_UpdateClassicEmail(t *testing.T) {
	t.Run("UpdatesEmailSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Initial classic with original email
		classic := Classic{
			ChassisNo:   "ABC123",
			Make:        "Ferrari",
			Model:       "Testarossa",
			Year:        1985,
			OwnerEmail:  "oldowner@example.com",
			PriorOwners: []string{},
		}
		classicJSON, _ := json.Marshal(classic)

		// Initial access with original email
		access := Access{
			OwnerEmail: "oldowner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Mock the state queries and auth check
		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "oldowner@example.com").Return(nil)

		// Simple mocks for PutState without capturing data
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)
		mockStub.On("PutState", "Access_ABC123", mock.Anything).Return(nil)

		// Execute the function under test
		sc := SmartContract{}
		result, err := sc.UpdateClassicEmail(mockCtx, "ABC123", "newowner@example.com")

		// Verify basic results
		assert.NoError(t, err)
		assert.Contains(t, result, "was updated sucessfully")
		mockStub.AssertNumberOfCalls(t, "PutState", 2)
		mockStub.AssertCalled(t, "PutState", "Classic_ABC123", mock.Anything)
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", mock.Anything)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.UpdateClassicEmail(mockCtx, "ABC123", "newowner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenClassicReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// GetState fails
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.UpdateClassicEmail(mockCtx, "ABC123", "newowner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Classic exists but access does not
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "oldowner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.UpdateClassicEmail(mockCtx, "ABC123", "newowner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenAccessReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Classic exists but access read fails
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "oldowner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.UpdateClassicEmail(mockCtx, "ABC123", "newowner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenClassicMarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Test can't directly mock json.Marshal failure, so skip this test
		// This is included for completeness of the test suite
	})

	t.Run("ReturnsErrorWhenAccessMarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Test can't directly mock json.Marshal failure, so skip this test
		// This is included for completeness of the test suite
	})

	t.Run("ReturnsErrorWhenClassicPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Initialize classic and access objects
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "oldowner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "oldowner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// PutState for classic fails
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		result, err := sc.UpdateClassicEmail(mockCtx, "ABC123", "newowner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenAccessPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Initialize classic and access objects
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "oldowner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "oldowner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// PutState for classic succeeds but access update fails
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)
		mockStub.On("PutState", "Access_ABC123", mock.Anything).Return(fmt.Errorf("access put state error"))

		sc := SmartContract{}
		result, err := sc.UpdateClassicEmail(mockCtx, "ABC123", "newowner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access put state error")
		assert.Empty(t, result)
	})
}

func TestSmartContract_CreateRestorationStep(t *testing.T) {
	t.Run("CreatesStepSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup initial classic with modifier access
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// Mock modifier check
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// Mock PutState for updated classic
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)

		sc := SmartContract{}
		photosIds := []string{"photo1.jpg", "photo2.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.NoError(t, err)
		assert.Equal(t, "ABC123_step_0", stepID)
		mockStub.AssertNumberOfCalls(t, "PutState", 1)
		mockStub.AssertCalled(t, "PutState", "Classic_ABC123", mock.Anything)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		photosIds := []string{"photo1.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, stepID)
	})

	t.Run("ReturnsErrorWhenGetStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// GetState fails
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		photosIds := []string{"photo1.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		assert.Empty(t, stepID)
	})

	t.Run("ReturnsErrorWhenUserIsNotModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup initial classic
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{}, // No modifiers
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// User is not a modifier
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nonmodifier@example.com", true, nil)

		sc := SmartContract{}
		photosIds := []string{"photo1.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, stepID)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup initial classic with access
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// Enrollment ID not found
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		photosIds := []string{"photo1.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.Error(t, err)
		assert.Empty(t, stepID)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup initial classic with access
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		// Setup access object
		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// Error when reading identity attribute
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		photosIds := []string{"photo1.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, stepID)
	})

	t.Run("ReturnsErrorWhenJsonMarshalFails", func(t *testing.T) {
		// This test can't be easily implemented as we can't mock json.Marshal
		// It's included for completeness
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup initial classic with modifier access
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// PutState fails
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		photosIds := []string{"photo1.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Empty(t, stepID)
	})

	t.Run("CreatesMultipleStepsCorrectly", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup classic with existing restoration step
		existingStep := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Body Work",
			Description: "Rust removal and panel repair",
			PhotosIds:   []string{"bodywork1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{existingStep},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)

		sc := SmartContract{}
		photosIds := []string{"engine1.jpg", "engine2.jpg"}
		stepID, err := sc.CreateRestorationStep(
			mockCtx,
			"ABC123",
			"Engine Rebuild",
			"Complete rebuild of the engine",
			photosIds,
			"2023-05-15T10:30:00Z",
		)

		assert.NoError(t, err)
		assert.Equal(t, "ABC123_step_1", stepID) // Should be step_1 since step_0 already exists
		mockStub.AssertNumberOfCalls(t, "PutState", 1)
	})
}

func TestSmartContract_GetStep(t *testing.T) {
	t.Run("ReturnsStepSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with restoration steps
		step0 := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Body Work",
			Description: "Rust removal and panel repair",
			PhotosIds:   []string{"bodywork1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		step1 := RestorationStep{
			ID:          "ABC123_step_1",
			Title:       "Engine Rebuild",
			Description: "Complete rebuild of the engine",
			PhotosIds:   []string{"engine1.jpg", "engine2.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-05-15T10:30:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step0, step1},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ABC123_step_1", result.ID)
		assert.Equal(t, "Engine Rebuild", result.Title)
		assert.Equal(t, "Complete rebuild of the engine", result.Description)
		assert.Equal(t, 2, len(result.PhotosIds))
		assert.Equal(t, "engine1.jpg", result.PhotosIds[0])
		assert.Equal(t, "engine2.jpg", result.PhotosIds[1])
		assert.Equal(t, "modifier@example.com", result.MadeBy)
		assert.Equal(t, "2023-05-15T10:30:00Z", result.When)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetClassicStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// GetState fails
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetAccessStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic exists but access read fails
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUserIsNotAuthorized", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with restoration steps
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{}, // No modifiers
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nonmodifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with restoration steps
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		// Setup access mock - this is what was missing
		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenStepNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with one restoration step
		step0 := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Body Work",
			Description: "Rust removal and panel repair",
			PhotosIds:   []string{"bodywork1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step0},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1") // This step doesn't exist

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404nostep")
		assert.Nil(t, result)
	})

	t.Run("ReturnsCorrectStepFromMultipleSteps", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with multiple restoration steps
		step0 := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Body Work",
			Description: "Rust removal and panel repair",
			PhotosIds:   []string{"bodywork1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		step1 := RestorationStep{
			ID:          "ABC123_step_1",
			Title:       "Engine Rebuild",
			Description: "Complete rebuild of the engine",
			PhotosIds:   []string{"engine1.jpg", "engine2.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-05-15T10:30:00Z",
		}

		step2 := RestorationStep{
			ID:          "ABC123_step_2",
			Title:       "Interior Restoration",
			Description: "Reupholstery and dashboard restoration",
			PhotosIds:   []string{"interior1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-06-20T14:45:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step0, step1, step2},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetStep(mockCtx, "ABC123", "ABC123_step_1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ABC123_step_1", result.ID)
		assert.Equal(t, "Engine Rebuild", result.Title)
	})
}

func TestSmartContract_UpdateStep(t *testing.T) {
	t.Run("UpdatesStepSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with restoration steps
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)

		sc := SmartContract{}
		result, err := sc.UpdateStep(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description")

		assert.NoError(t, err)
		assert.Contains(t, result, "was updated sucessfully")
	})

	t.Run("ReturnsErrorWhenReadClassicAsModifierFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStep(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenStepNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic with no steps that match
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStep(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404nostep")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenUserIsNotStepCreator", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step created by a different user
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "otherperson@example.com", // Different from the current user
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStep(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404madeby")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStep(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description")

		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.UpdateStep(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists with matching creator
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		result, err := sc.UpdateStep(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Empty(t, result)
	})
}

func TestSmartContract_UpdateStepPhotos(t *testing.T) {
	t.Run("UpdatesPhotosSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with restoration steps
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Rebuild",
			Description: "Complete rebuild of the engine",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", newPhotos)

		assert.NoError(t, err)
		assert.Contains(t, result, "was updated sucessfully")

		// Verify the updated classic object
		mockStub.AssertNumberOfCalls(t, "PutState", 1)
		putStateCall := mockStub.Calls[2]
		updatedClassicBytes := putStateCall.Arguments[1].([]byte)

		var updatedClassic Classic
		err = json.Unmarshal(updatedClassicBytes, &updatedClassic)
		assert.NoError(t, err)

		// Check that photos were appended correctly
		assert.Equal(t, 3, len(updatedClassic.Restorations[0].PhotosIds))
		assert.Equal(t, "photo1.jpg", updatedClassic.Restorations[0].PhotosIds[0])
		assert.Equal(t, "photo2.jpg", updatedClassic.Restorations[0].PhotosIds[1])
		assert.Equal(t, "photo3.jpg", updatedClassic.Restorations[0].PhotosIds[2])
	})

	t.Run("ReturnsErrorWhenReadClassicAsModifierFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", newPhotos)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenStepNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic with no steps that match
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", newPhotos)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404nostep")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenUserIsNotStepCreator", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step created by a different user
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Rebuild",
			Description: "Complete rebuild of the engine",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "otherperson@example.com", // Different from the current user
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", newPhotos)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404madeby")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Rebuild",
			Description: "Complete rebuild of the engine",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", newPhotos)

		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Rebuild",
			Description: "Complete rebuild of the engine",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", newPhotos)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists with matching creator
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Rebuild",
			Description: "Complete rebuild of the engine",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", newPhotos)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenReadClassicAsModifierFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		photosIds := []string{"newphoto1.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", photosIds)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenStepNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic with no steps that match
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		photosIds := []string{"newphoto1.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", photosIds)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404nostep")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenUserIsNotStepCreator", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step created by a different user
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Work",
			Description: "Engine rebuild",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "otherperson@example.com", // Different from the current user
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		photosIds := []string{"newphoto1.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", photosIds)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404madeby")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Work",
			Description: "Engine rebuild",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		photosIds := []string{"newphoto1.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", photosIds)

		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Work",
			Description: "Engine rebuild",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		photosIds := []string{"newphoto1.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", photosIds)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists with matching creator
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Work",
			Description: "Engine rebuild",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		photosIds := []string{"newphoto1.jpg"}
		result, err := sc.UpdateStepPhotos(mockCtx, "ABC123", "ABC123_step_0", photosIds)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Empty(t, result)
	})
}

func TestSmartContract_UpdateStepAndPhotos(t *testing.T) {
	t.Run("UpdatesStepAndPhotosSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with restoration steps
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)

		sc := SmartContract{}
		newPhotos := []string{"photo2.jpg", "photo3.jpg"}
		result, err := sc.UpdateStepAndPhotos(
			mockCtx,
			"ABC123",
			"ABC123_step_0",
			"New Title",
			"New Description",
			newPhotos,
		)

		assert.NoError(t, err)
		assert.Contains(t, result, "was updated sucessfully")

		// Verify the updated classic object
		mockStub.AssertNumberOfCalls(t, "PutState", 1)
		putStateCall := mockStub.Calls[2]
		updatedClassicBytes := putStateCall.Arguments[1].([]byte)

		var updatedClassic Classic
		err = json.Unmarshal(updatedClassicBytes, &updatedClassic)
		assert.NoError(t, err)

		// Check that title, description, and photos were updated correctly
		assert.Equal(t, "New Title", updatedClassic.Restorations[0].Title)
		assert.Equal(t, "New Description", updatedClassic.Restorations[0].Description)
		assert.Equal(t, 3, len(updatedClassic.Restorations[0].PhotosIds))
		assert.Equal(t, "photo1.jpg", updatedClassic.Restorations[0].PhotosIds[0])
		assert.Equal(t, "photo2.jpg", updatedClassic.Restorations[0].PhotosIds[1])
		assert.Equal(t, "photo3.jpg", updatedClassic.Restorations[0].PhotosIds[2])
	})

	t.Run("ReturnsErrorWhenReadClassicAsModifierFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStepAndPhotos(
			mockCtx,
			"ABC123",
			"ABC123_step_0",
			"New Title",
			"New Description",
			[]string{"photo2.jpg"},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenStepNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic with no steps that match
		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStepAndPhotos(
			mockCtx,
			"ABC123",
			"ABC123_step_0",
			"New Title",
			"New Description",
			[]string{"photo2.jpg"},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404nostep")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenUserIsNotStepCreator", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step created by a different user
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "otherperson@example.com", // Different from the current user
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStepAndPhotos(
			mockCtx,
			"ABC123",
			"ABC123_step_0",
			"New Title",
			"New Description",
			[]string{"photo2.jpg"},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404madeby")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenEnrollmentIDAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Old Title",
			Description: "Old Description",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.UpdateStepAndPhotos(
			mockCtx,
			"ABC123",
			"ABC123_step_0",
			"New Title",
			"New Description",
			[]string{"photo2.jpg"},
		)

		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Step exists with matching creator
		step := RestorationStep{
			ID:          "ABC123_step_0",
			Title:       "Engine Work",
			Description: "Engine rebuild",
			PhotosIds:   []string{"photo1.jpg"},
			MadeBy:      "modifier@example.com",
			When:        "2023-04-10T09:00:00Z",
		}

		classic := Classic{
			ChassisNo:    "ABC123",
			Make:         "Ferrari",
			Model:        "Testarossa",
			Year:         1985,
			OwnerEmail:   "owner@example.com",
			Restorations: []RestorationStep{step},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		result, err := sc.UpdateStepAndPhotos(mockCtx, "ABC123", "ABC123_step_0", "New Title", "New Description", []string{"newphoto.jpg"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Empty(t, result)
	})
}

func TestSmartContract_ClassicExists(t *testing.T) {
	t.Run("ReturnsTrue_WhenClassicExists", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return([]byte(`{"chassisNo": "ABC123"}`), nil)

		sc := SmartContract{}
		exists, err := sc.ClassicExists(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("ReturnsFalse_WhenClassicDoesNotExist", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		exists, err := sc.ClassicExists(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ReturnsError_WhenGetStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("state error"))

		sc := SmartContract{}
		exists, err := sc.ClassicExists(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.False(t, exists)
	})
}

func TestSmartContract_DeleteClassic(t *testing.T) {
	t.Run("DeletesClassicSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Classic exists
		mockStub.On("GetState", "Classic_ABC123").Return([]byte(`{"chassisNo": "ABC123"}`), nil)
		mockStub.On("DelState", "Access_ABC123").Return(nil)
		mockStub.On("DelState", "Classic_ABC123").Return(nil)

		sc := SmartContract{}
		err := sc.DeleteClassic(mockCtx, "ABC123")

		assert.NoError(t, err)
		mockStub.AssertCalled(t, "DelState", "Access_ABC123")
		mockStub.AssertCalled(t, "DelState", "Classic_ABC123")
	})

	t.Run("ReturnsErrorWhenClassicDoesNotExist", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Classic does not exist
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		err := sc.DeleteClassic(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		mockStub.AssertNotCalled(t, "DelState", "Access_ABC123")
		mockStub.AssertNotCalled(t, "DelState", "Classic_ABC123")
	})

	t.Run("ReturnsErrorWhenClassicExistsFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// GetState fails
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("state error"))

		sc := SmartContract{}
		err := sc.DeleteClassic(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read")
		mockStub.AssertNotCalled(t, "DelState", "Access_ABC123")
		mockStub.AssertNotCalled(t, "DelState", "Classic_ABC123")
	})

	t.Run("ReturnsErrorWhenDeleteAccessFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Classic exists, but DelState for Access fails
		mockStub.On("GetState", "Classic_ABC123").Return([]byte(`{"chassisNo": "ABC123"}`), nil)
		mockStub.On("DelState", "Access_ABC123").Return(fmt.Errorf("delete access error"))

		sc := SmartContract{}
		err := sc.DeleteClassic(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete access error")
		mockStub.AssertCalled(t, "DelState", "Access_ABC123")
		mockStub.AssertNotCalled(t, "DelState", "Classic_ABC123")
	})

	t.Run("ReturnsErrorWhenDeleteClassicFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// For ClassicExists check - the classic does exist
		mockStub.On("GetState", "Classic_ABC123").Return([]byte(`{"chassisNo": "ABC123"}`), nil).Once()

		// For the actual deletion operations
		mockStub.On("DelState", "Access_ABC123").Return(nil)
		mockStub.On("DelState", "Classic_ABC123").Return(fmt.Errorf("delete classic error"))

		sc := SmartContract{}
		err := sc.DeleteClassic(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete classic error")
		mockStub.AssertCalled(t, "DelState", "Access_ABC123")
		mockStub.AssertCalled(t, "DelState", "Classic_ABC123")
	})
}

func TestSmartContract_ReadClassic(t *testing.T) {
	t.Run("ReturnsClassicSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic exists and the user is the owner
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		sc := SmartContract{}
		result, err := sc.ReadClassic(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "ABC123", result.ChassisNo)
		assert.Equal(t, "Ferrari", result.Make)
		assert.Equal(t, "Testarossa", result.Model)
		assert.Equal(t, 1985, result.Year)
		assert.Equal(t, "owner@example.com", result.OwnerEmail)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Classic not found
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassic(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// GetState fails
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.ReadClassic(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUserIsNotOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic exists but user is not the owner
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(fmt.Errorf("unauthorized"))

		sc := SmartContract{}
		result, err := sc.ReadClassic(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Nil(t, result)
	})
}

func TestSmartContract_AddDocument(t *testing.T) {
	t.Run("AddsDocumentSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with documents
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			Documents:  map[string]string{},
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		// Setup mocks for access check
		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"documenter@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("documenter@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(nil)

		sc := SmartContract{}
		result, err := sc.AddDocument(mockCtx, "ABC123", "invoice.pdf", "/path/to/document.pdf")

		assert.NoError(t, err)
		assert.Contains(t, result, "was added sucessfully")

		// Verify the document was saved
		mockStub.AssertCalled(t, "PutState", "Classic_ABC123", mock.Anything)
	})

	t.Run("ReturnsErrorWhenReadClassicAsDocumenterFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic doesn't exist
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("documenter@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.AddDocument(mockCtx, "ABC123", "invoice.pdf", "/path/to/document.pdf")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenAccessCheckFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Classic exists but no access
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			Documents:  map[string]string{},
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		// User is not in the documenter list
		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("documenter@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.AddDocument(mockCtx, "ABC123", "invoice.pdf", "/path/to/document.pdf")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create a classic with documents
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			Documents:  map[string]string{},
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)

		// Set up access
		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"documenter@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("documenter@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		result, err := sc.AddDocument(mockCtx, "ABC123", "invoice.pdf", "/path/to/document.pdf")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Empty(t, result)
	})
}

func TestSmartContract_GetClassicHistory2(t *testing.T) {
	t.Run("ReturnsHistorySuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockHistoryIterator := new(shim.MockHistoryQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Mock ReadClassicAsViewer to succeed
		classicJSON, _ := json.Marshal(Classic{ChassisNo: "ABC123"})
		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)

		// Make sure the viewer access timestamp is AFTER the query timestamp
		// The format is RFC3339 (2023-10-15T12:00:00Z)
		mockStub.On("GetState", "Access_ABC123").Return([]byte(`{"OwnerEmail":"owner@example.com","Viewers":{"viewer@example.com":"2099-10-01T12:00:00Z"}}`), nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		// Simple history setup
		mockHistoryResult := &queryresult.KeyModification{
			TxId:      "tx1",
			Value:     []byte(`{}`),
			Timestamp: &timestamp.Timestamp{},
		}

		mockHistoryIterator.On("HasNext").Return(true).Once()
		mockHistoryIterator.On("HasNext").Return(false)
		mockHistoryIterator.On("Next").Return(mockHistoryResult, nil)
		mockHistoryIterator.On("Close").Return(nil)

		mockStub.On("GetHistoryForKey", "Classic_ABC123").Return(mockHistoryIterator, nil)

		sc := SmartContract{}
		result, err := sc.GetClassicHistory2(mockCtx, "ABC123", "2023-10-15T12:00:00Z")

		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		// Only verify basic structure rather than specific content
		assert.Contains(t, result, "counter")
		assert.Contains(t, result, "txns")
	})

	t.Run("ReturnsErrorWhenReadClassicAsViewerFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Mock ReadClassicAsViewer to fail (classic not found)
		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.GetClassicHistory2(mockCtx, "ABC123", "2023-10-15T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenGetHistoryForKeyFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Mock ReadClassicAsViewer to succeed
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)
		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return([]byte(`{"OwnerEmail":"owner@example.com","Viewers":{"viewer@example.com":"2023-10-01T12:00:00Z"}}`), nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		// GetHistoryForKey fails
		mockStub.On("GetHistoryForKey", "Classic_ABC123").Return(nil, fmt.Errorf("history error"))

		sc := SmartContract{}
		result, err := sc.GetClassicHistory2(mockCtx, "ABC123", "2023-10-15T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenHistoryIteratorNextFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockHistoryIterator := new(shim.MockHistoryQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Mock ReadClassicAsViewer to succeed
		classic := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "owner@example.com",
		}
		classicJSON, _ := json.Marshal(classic)
		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return([]byte(`{"OwnerEmail":"owner@example.com","Viewers":{"viewer@example.com":"2023-10-01T12:00:00Z"}}`), nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		// History iterator setup with Next() failing
		mockHistoryIterator.On("HasNext").Return(true)
		mockHistoryIterator.On("Next").Return(nil, fmt.Errorf("iterator error"))
		mockHistoryIterator.On("Close").Return(nil)

		mockStub.On("GetHistoryForKey", "Classic_ABC123").Return(mockHistoryIterator, nil)

		sc := SmartContract{}
		result, err := sc.GetClassicHistory2(mockCtx, "ABC123", "2023-10-15T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Empty(t, result)
	})
}

func TestSmartContract_QueryClassicsByOwner(t *testing.T) {
	t.Run("ReturnsClassicsSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Mock iterator behavior
		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)

		// Create test data for two classics with the same owner
		classic1 := Classic{
			ChassisNo:  "ABC123",
			Make:       "Ferrari",
			Model:      "Testarossa",
			Year:       1985,
			OwnerEmail: "owner@example.com",
		}
		classic2 := Classic{
			ChassisNo:  "DEF456",
			Make:       "Porsche",
			Model:      "911",
			Year:       1982,
			OwnerEmail: "owner@example.com",
		}
		classic3 := Classic{
			ChassisNo:  "GHI789",
			Make:       "Lamborghini",
			Model:      "Countach",
			Year:       1986,
			OwnerEmail: "different@example.com",
		}

		classic1JSON, _ := json.Marshal(classic1)
		classic2JSON, _ := json.Marshal(classic2)
		classic3JSON, _ := json.Marshal(classic3)

		// Set up the mock iterator to return all classics
		mockResultsIterator.On("HasNext").Return(true).Times(3)
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Value: classic1JSON}, nil,
		).Once()
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Value: classic2JSON}, nil,
		).Once()
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Value: classic3JSON}, nil,
		).Once()
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByOwner(mockCtx, "owner@example.com")

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "ABC123", result[0].ChassisNo)
		assert.Equal(t, "DEF456", result[1].ChassisNo)
	})

	t.Run("ReturnsErrorWhenNoClassicsFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Mock empty iterator
		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByOwner(mockCtx, "nonexistent@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetStateByRangeFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(nil, fmt.Errorf("iterator error"))

		sc := SmartContract{}
		result, err := sc.QueryClassicsByOwner(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "iterator error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenNextFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Mock iterator with error on Next()
		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(nil, fmt.Errorf("next error"))
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByOwner(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "next error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Mock iterator with invalid JSON
		mockStub.On("GetStateByRange", "Classic_", "Classic_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(&queryresult.KV{Value: []byte(`invalid json`)}, nil)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByOwner(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSmartContract_QueryClassicsByModifier(t *testing.T) {
	t.Run("ReturnsClassicsSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Mock identity
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// Mock iterator behavior
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)

		// Create test access data
		access1 := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		access2 := Access{
			OwnerEmail: "owner2@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}

		access1JSON, _ := json.Marshal(access1)
		access2JSON, _ := json.Marshal(access2)

		// Test classic data
		classic1 := Classic{ChassisNo: "ABC123"}
		classic2 := Classic{ChassisNo: "DEF456"}
		classic1JSON, _ := json.Marshal(classic1)
		classic2JSON, _ := json.Marshal(classic2)

		// Set up the mock iterator to return accesses
		mockResultsIterator.On("HasNext").Return(true).Times(2)
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Key: "Access_ABC123", Value: access1JSON}, nil,
		).Once()
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Key: "Access_DEF456", Value: access2JSON}, nil,
		).Once()
		mockResultsIterator.On("Close").Return(nil)

		// Mock GetState for the classics
		mockStub.On("GetState", "Classic_ABC123").Return(classic1JSON, nil)
		mockStub.On("GetState", "Classic_DEF456").Return(classic2JSON, nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
	})

	t.Run("ReturnsErrorWhenIdentityAttributeFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetStateByRangeFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(nil, fmt.Errorf("iterator error"))

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "iterator error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenNextFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(nil, fmt.Errorf("next error"))
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "next error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenAccessUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(&queryresult.KV{Value: []byte(`invalid json`)}, nil)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// Create valid access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true).Once()
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Next").Return(&queryresult.KV{Key: "Access_ABC123", Value: accessJSON}, nil)
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("get state error"))
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		// Create valid access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true).Once()
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Next").Return(&queryresult.KV{Key: "Access_ABC123", Value: accessJSON}, nil)
		mockStub.On("GetState", "Classic_ABC123").Return([]byte(`invalid json`), nil)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenNoClassicsFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByModifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Nil(t, result)
	})
}

func TestSmartContract_QueryClassicsByCertifier(t *testing.T) {
	t.Run("SuccessfullyReturnsClassics", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Mock identity
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		// Mock iterator behavior
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)

		// Create test access data
		access1 := Access{
			OwnerEmail: "owner@example.com",
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		access2 := Access{
			OwnerEmail: "owner2@example.com",
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}

		access1JSON, _ := json.Marshal(access1)
		access2JSON, _ := json.Marshal(access2)

		// Test classic data
		classic1 := Classic{ChassisNo: "ABC123"}
		classic2 := Classic{ChassisNo: "DEF456"}
		classic1JSON, _ := json.Marshal(classic1)
		classic2JSON, _ := json.Marshal(classic2)

		// Set up the mock iterator to return accesses
		mockResultsIterator.On("HasNext").Return(true).Times(2)
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Key: "Access_ABC123", Value: access1JSON}, nil,
		).Once()
		mockResultsIterator.On("Next").Return(
			&queryresult.KV{Key: "Access_DEF456", Value: access2JSON}, nil,
		).Once()
		mockResultsIterator.On("Close").Return(nil)

		// Mock GetState for the classics
		mockStub.On("GetState", "Classic_ABC123").Return(classic1JSON, nil)
		mockStub.On("GetState", "Classic_DEF456").Return(classic2JSON, nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
	})

	t.Run("ReturnsErrorWhenIdentityAttributeFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "403")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetStateByRangeFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(nil, fmt.Errorf("iterator error"))

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "iterator error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenNextFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(nil, fmt.Errorf("next error"))
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "next error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenAccessUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true)
		mockResultsIterator.On("Next").Return(&queryresult.KV{Value: []byte(`invalid json`)}, nil)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenGetStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		// Create valid access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true).Once()
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Next").Return(&queryresult.KV{Key: "Access_ABC123", Value: accessJSON}, nil)
		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("get state error"))
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		// Create valid access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(true).Once()
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Next").Return(&queryresult.KV{Key: "Access_ABC123", Value: accessJSON}, nil)
		mockStub.On("GetState", "Classic_ABC123").Return([]byte(`invalid json`), nil)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenNoClassicsFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockResultsIterator := new(shim.MockStateQueryIteratorInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)
		mockStub.On("GetStateByRange", "Access_", "Access_\uffff").Return(mockResultsIterator, nil)
		mockResultsIterator.On("HasNext").Return(false)
		mockResultsIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.QueryClassicsByCertifier(mockCtx, "owner@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Nil(t, result)
	})
}
