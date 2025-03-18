package v1

import (
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/queryresult"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

func TestSmartContract_GetAccess(t *testing.T) {
	t.Run("ReturnsAccessDataWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2022-01-01T00:00:00Z"},
			Modifiers:  map[string]string{"modifier@example.com": "2022-01-01T00:00:00Z"},
			Certifiers: map[string]string{"certifier@example.com": "2022-01-01T00:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAccess(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.Equal(t, &access, result)
	})

	t.Run("Returns404WhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		_, err := sc.GetAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		_, err := sc.GetAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		_, err := sc.GetAccess(mockCtx, "ABC123")

		assert.Error(t, err)
	})

	t.Run("Returns403WhenUserIsNotOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(fmt.Errorf("user is not owner"))

		sc := SmartContract{}
		_, err := sc.GetAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
	})

	t.Run("UsesCorrectKeyFormat", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_XYZ789").Return(nil, nil)

		sc := SmartContract{}
		_, _ = sc.GetAccess(mockCtx, "XYZ789")

		mockStub.AssertCalled(t, "GetState", "Access_XYZ789")
	})
}

func TestSmartContract_GiveAccess(t *testing.T) {
	t.Run("AddsViewerAccessSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Initial access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Expected updated access with new viewer
		expectedAccess := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"newviewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		expectedJSON, _ := json.Marshal(expectedAccess)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", expectedJSON).Return(nil)

		sc := SmartContract{}
		result, err := sc.GiveAccess(mockCtx, "ABC123", "newviewer@example.com", "viewer", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Contains(t, result, "access permissions were updated")
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", expectedJSON)
	})

	t.Run("AddsModifierAccessSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		expectedAccess := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		expectedJSON, _ := json.Marshal(expectedAccess)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", expectedJSON).Return(nil)

		sc := SmartContract{}
		result, err := sc.GiveAccess(mockCtx, "ABC123", "modifier@example.com", "modifier", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Contains(t, result, "access permissions were updated")
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", expectedJSON)
	})

	t.Run("AddsCertifierAccessSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		expectedAccess := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		expectedJSON, _ := json.Marshal(expectedAccess)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", expectedJSON).Return(nil)

		sc := SmartContract{}
		result, err := sc.GiveAccess(mockCtx, "ABC123", "certifier@example.com", "certifier", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Contains(t, result, "access permissions were updated")
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", expectedJSON)
	})

	t.Run("ReturnsErrorForInvalidAccessLevel", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		sc := SmartContract{}
		_, err := sc.GiveAccess(mockCtx, "ABC123", "someone@example.com", "invalid", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "404nolevel", err.Error())
	})

	t.Run("ReturnsErrorWhenNotOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(fmt.Errorf("not owner"))

		sc := SmartContract{}
		_, err := sc.GiveAccess(mockCtx, "ABC123", "someone@example.com", "viewer", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", mock.Anything).Return(fmt.Errorf("put state failed"))

		sc := SmartContract{}
		_, err := sc.GiveAccess(mockCtx, "ABC123", "someone@example.com", "viewer", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state failed")
	})
}

func TestSmartContract_RevokeAccess(t *testing.T) {
	t.Run("RevokesViewerAccessSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Initial access data with viewer
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Expected updated access after revoking
		expectedAccess := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		expectedJSON, _ := json.Marshal(expectedAccess)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", expectedJSON).Return(nil)

		sc := SmartContract{}
		result, err := sc.RevokeAccess(mockCtx, "ABC123", "viewer@example.com", "viewer")

		assert.NoError(t, err)
		assert.Contains(t, result, "access permissions revoked")
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", expectedJSON)
	})

	t.Run("RevokesModifierAccessSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		expectedAccess := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		expectedJSON, _ := json.Marshal(expectedAccess)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", expectedJSON).Return(nil)

		sc := SmartContract{}
		result, err := sc.RevokeAccess(mockCtx, "ABC123", "modifier@example.com", "modifier")

		assert.NoError(t, err)
		assert.Contains(t, result, "access permissions revoked")
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", expectedJSON)
	})

	t.Run("RevokesCertifierAccessSuccessfully", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		expectedAccess := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		expectedJSON, _ := json.Marshal(expectedAccess)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", expectedJSON).Return(nil)

		sc := SmartContract{}
		result, err := sc.RevokeAccess(mockCtx, "ABC123", "certifier@example.com", "certifier")

		assert.NoError(t, err)
		assert.Contains(t, result, "access permissions revoked")
		mockStub.AssertCalled(t, "PutState", "Access_ABC123", expectedJSON)
	})

	t.Run("IgnoresNonExistingUserAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Expected should be the same as original since no change occurred
		expectedJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", expectedJSON).Return(nil)

		sc := SmartContract{}
		result, err := sc.RevokeAccess(mockCtx, "ABC123", "nonexisting@example.com", "viewer")

		assert.NoError(t, err)
		assert.Contains(t, result, "access permissions revoked")
	})

	t.Run("ReturnsErrorForInvalidAccessLevel", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		sc := SmartContract{}
		_, err := sc.RevokeAccess(mockCtx, "ABC123", "someone@example.com", "invalid")

		assert.Error(t, err)
		assert.Equal(t, "404nolevel", err.Error())
	})

	t.Run("ReturnsErrorWhenNotOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(fmt.Errorf("not owner"))

		sc := SmartContract{}
		_, err := sc.RevokeAccess(mockCtx, "ABC123", "viewer@example.com", "viewer")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)
		mockStub.On("PutState", "Access_ABC123", mock.Anything).Return(fmt.Errorf("put state failed"))

		sc := SmartContract{}
		_, err := sc.RevokeAccess(mockCtx, "ABC123", "viewer@example.com", "viewer")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state failed")
	})
}

func TestSmartContract_HasViewerAccess(t *testing.T) {
	t.Run("ReturnsTrueWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("owner@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsTrueWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsTrueWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsTrueWhenUserIsViewerWithinTimeLimit", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Access granted for 2023-01-01T12:00:00Z, checked within 24h
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		// Current time is 12 hours after access was granted
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-02T00:00:00Z")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsFalseWhenUserIsViewerButExpired", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Access granted for 2023-01-01T12:00:00Z, checked after 24h
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		// Current time is 36 hours after access was granted (exceeds 24h limit)
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-03T00:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.HasViewerAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.False(t, result)
	})
}

func TestSmartContract_ReadClassicAsViewer(t *testing.T) {
	t.Run("ReturnsClassicWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("owner@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsClassicWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsClassicWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsClassicWhenUserIsViewerWithinTimeLimit", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-02T00:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsErrorWhenUserIsViewerButExpired", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-03T00:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsViewer(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSmartContract_HasModifierAccess(t *testing.T) {
	t.Run("ReturnsTrueWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("owner@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsTrueWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsFalseWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsFalseWhenUserIsViewer", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.HasModifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})

}

func TestSmartContract_ReadClassicAsModifier(t *testing.T) {
	t.Run("ReturnsClassicWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("owner@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsClassicWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsErrorWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUserIsViewer", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsModifier(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSmartContract_HasCertifierAccess(t *testing.T) {
	t.Run("ReturnsTrueWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsFalseWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsFalseWhenUserIsViewer", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.HasCertifierAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})
}

func TestSmartContract_MarkAsCertified(t *testing.T) {
	t.Run("CertifiesClassicWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo:      "ABC123",
			Certifications: []string{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.MatchedBy(func(data []byte) bool {
			var updatedClassic Classic
			json.Unmarshal(data, &updatedClassic)
			return len(updatedClassic.Certifications) == 1 && updatedClassic.Certifications[0] == "certifier@example.com"
		})).Return(nil)

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Certifications))
		assert.Equal(t, "certifier@example.com", result.Certifications[0])
		mockStub.AssertCalled(t, "PutState", "Classic_ABC123", mock.Anything)
	})

	t.Run("AppendsToExistingCertifications", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo:      "ABC123",
			Certifications: []string{"existing@example.com"},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.MatchedBy(func(data []byte) bool {
			var updatedClassic Classic
			json.Unmarshal(data, &updatedClassic)
			return len(updatedClassic.Certifications) == 2 &&
				updatedClassic.Certifications[0] == "existing@example.com" &&
				updatedClassic.Certifications[1] == "certifier@example.com"
		})).Return(nil)

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result.Certifications))
		assert.Equal(t, "existing@example.com", result.Certifications[0])
		assert.Equal(t, "certifier@example.com", result.Certifications[1])
		mockStub.AssertCalled(t, "PutState", "Classic_ABC123", mock.Anything)
	})

	t.Run("ReturnsErrorWhenUserIsNotCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo:      "ABC123",
			Certifications: []string{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenPutStateFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo:      "ABC123",
			Certifications: []string{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)
		mockStub.On("PutState", "Classic_ABC123", mock.Anything).Return(fmt.Errorf("put state error"))

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put state error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo:      "ABC123",
			Certifications: []string{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// First call in HasCertifierAccess returns success
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil).Once()

		// Second call in MarkAsCertified fails
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error")).Once()

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "identity error")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFoundInCertificationStep", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo:      "ABC123",
			Certifications: []string{},
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)

		// First call in HasCertifierAccess returns success
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil).Once()

		// Second call in MarkAsCertified returns attribute not found
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil).Once()

		sc := SmartContract{}
		result, err := sc.MarkAsCertified(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSmartContract_HasDocumenterAccess(t *testing.T) {
	t.Run("ReturnsTrueWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("owner@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsTrueWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsTrueWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("ReturnsFalseWhenUserIsViewer", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.HasDocumenterAccess(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "identity error")
		assert.False(t, result)
	})
}

func TestSmartContract_ReadClassicAsDocumenter(t *testing.T) {
	t.Run("ReturnsClassicWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("owner@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsClassicWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsClassicWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.Equal(t, &classic, result)
	})

	t.Run("ReturnsErrorWhenUserIsViewer", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenClassicNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Classic_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("ReturnsErrorWhenAccessCheckFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		classic := Classic{
			ChassisNo: "ABC123",
			Make:      "Ferrari",
			Model:     "Testarossa",
			Year:      1985,
		}
		classicJSON, _ := json.Marshal(classic)

		mockStub.On("GetState", "Classic_ABC123").Return(classicJSON, nil)
		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("access check error"))

		sc := SmartContract{}
		result, err := sc.ReadClassicAsDocumenter(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSmartContract_CheckUserAccess(t *testing.T) {
	t.Run("ReturnsOwnerWhenUserIsOwner", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("owner@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, accessLevelOwner, result)
	})

	t.Run("ReturnsModifierWhenUserIsModifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{"modifier@example.com": "2023-01-01T12:00:00Z"},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("modifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, accessLevelModifier, result)
	})

	t.Run("ReturnsCertifierWhenUserIsCertifier", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{"certifier@example.com": "2023-01-01T12:00:00Z"},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("certifier@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, accessLevelCertifier, result)
	})

	t.Run("ReturnsViewerWhenUserIsViewerWithinTimeLimit", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		// Test within 24h limit (1440 minutes)
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-02T00:00:00Z")

		assert.NoError(t, err)
		assert.Equal(t, accessLevelViewer, result)
	})

	t.Run("ReturnsErrorWhenUserIsViewerButExpired", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		// Test beyond 24h limit (1440 minutes)
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-03T00:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Equal(t, "", result)
	})

	t.Run("ReturnsErrorForInvalidTimeFormat", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "invalid-time-format"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("viewer@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("ReturnsErrorWhenUserHasNoAccess", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("nouser@example.com", true, nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Equal(t, "", result)
	})

	t.Run("ReturnsErrorWhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
		assert.Equal(t, "", result)
	})

	t.Run("ReturnsErrorWhenWorldStateReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Equal(t, "", result)
	})

	t.Run("ReturnsErrorWhenUnmarshalFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		mockStub.On("GetState", "Access_ABC123").Return([]byte("invalid json"), nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeReadFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, fmt.Errorf("identity error"))

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "identity error")
		assert.Equal(t, "", result)
	})

	t.Run("ReturnsErrorWhenIdentityAttributeNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("GetAttributeValue", enrollmentIDAtt).Return("", false, nil)

		sc := SmartContract{}
		result, err := sc.CheckUserAccess(mockCtx, "ABC123", "2023-01-01T12:00:00Z")

		assert.Error(t, err)
		assert.Equal(t, "", result)
	})
}

func TestSmartContract_GetAccessHistory(t *testing.T) {
	t.Run("ReturnsHistoryWhenAccessExists", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockHistoryIterator := new(shim.MockHistoryQueryIteratorInterface)

		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create mock access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Setup for GetAccess
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		// Setup for GetHistoryForKey
		mockStub.On("GetHistoryForKey", "Access_ABC123").Return(mockHistoryIterator, nil)

		// Create mock history responses
		timestamp1 := &timestamppb.Timestamp{Seconds: 1609459200} // 2021-01-01
		timestamp2 := &timestamppb.Timestamp{Seconds: 1609545600} // 2021-01-02

		mockHistoryResp1 := &queryresult.KeyModification{
			TxId:      "tx1",
			Value:     accessJSON,
			Timestamp: timestamp1,
		}

		// Updated access with a viewer
		updatedAccess := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{"viewer@example.com": "2023-01-01T12:00:00Z"},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		updatedAccessJSON, _ := json.Marshal(updatedAccess)

		mockHistoryResp2 := &queryresult.KeyModification{
			TxId:      "tx2",
			Value:     updatedAccessJSON,
			Timestamp: timestamp2,
		}

		mockHistoryIterator.On("HasNext").Return(true).Once()
		mockHistoryIterator.On("HasNext").Return(true).Once()
		mockHistoryIterator.On("HasNext").Return(false)
		mockHistoryIterator.On("Next").Return(mockHistoryResp1, nil).Once()
		mockHistoryIterator.On("Next").Return(mockHistoryResp2, nil).Once()
		mockHistoryIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAccessHistory(mockCtx, "ABC123")

		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Test behavior rather than exact JSON structure
		var historyData map[string]interface{}
		err = json.Unmarshal([]byte(result), &historyData)
		assert.NoError(t, err)

		// Check counter
		counter, exists := historyData["counter"]
		assert.True(t, exists)
		assert.Equal(t, float64(2), counter)

		// Check transactions exist
		txns, exists := historyData["txns"]
		assert.True(t, exists)
		txnsArray, ok := txns.([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(txnsArray))

		// Check first transaction
		tx1, ok := txnsArray[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "tx1", tx1["txn"])

		// Check second transaction
		tx2, ok := txnsArray[1].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "tx2", tx2["txn"])
	})

	t.Run("ReturnsErrorWhenAccessNotFound", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Setup for GetAccess - access not found
		mockStub.On("GetState", "Access_ABC123").Return(nil, nil)

		sc := SmartContract{}
		result, err := sc.GetAccessHistory(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "404", err.Error())
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenGetAccessFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockCtx.On("GetStub").Return(mockStub)

		// Setup for GetAccess - db error
		mockStub.On("GetState", "Access_ABC123").Return(nil, fmt.Errorf("database error"))

		sc := SmartContract{}
		result, err := sc.GetAccessHistory(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read from world state")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenAuthorizationFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create mock access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Setup for GetAccess but authorization fails
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(fmt.Errorf("403"))

		sc := SmartContract{}
		result, err := sc.GetAccessHistory(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Equal(t, "403", err.Error())
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenHistoryIteratorCreationFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create mock access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Setup for GetAccess
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		// Setup for GetHistoryForKey - failure
		mockStub.On("GetHistoryForKey", "Access_ABC123").Return(nil, fmt.Errorf("history error"))

		sc := SmartContract{}
		result, err := sc.GetAccessHistory(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get access permissions history")
		assert.Empty(t, result)
	})

	t.Run("ReturnsErrorWhenNextFails", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockHistoryIterator := new(shim.MockHistoryQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create mock access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Setup for GetAccess
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		// Setup for GetHistoryForKey
		mockStub.On("GetHistoryForKey", "Access_ABC123").Return(mockHistoryIterator, nil)

		mockHistoryIterator.On("HasNext").Return(true).Once()
		mockHistoryIterator.On("Next").Return(nil, fmt.Errorf("iterator error"))
		mockHistoryIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAccessHistory(mockCtx, "ABC123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "iterator error")
		assert.Empty(t, result)
	})

	t.Run("ReturnsEmptyHistoryWhenNoTransactions", func(t *testing.T) {
		mockCtx := new(contractapi.MockTransactionContextInterface)
		mockStub := new(shim.MockChaincodeStubInterface)
		mockIdentity := new(cid.MockClientIdentity)
		mockHistoryIterator := new(shim.MockHistoryQueryIteratorInterface)
		mockCtx.On("GetStub").Return(mockStub)
		mockCtx.On("GetClientIdentity").Return(mockIdentity)

		// Create mock access data
		access := Access{
			OwnerEmail: "owner@example.com",
			Viewers:    map[string]string{},
			Modifiers:  map[string]string{},
			Certifiers: map[string]string{},
		}
		accessJSON, _ := json.Marshal(access)

		// Setup for GetAccess
		mockStub.On("GetState", "Access_ABC123").Return(accessJSON, nil)
		mockIdentity.On("AssertAttributeValue", enrollmentIDAtt, "owner@example.com").Return(nil)

		// Setup for GetHistoryForKey - no history
		mockStub.On("GetHistoryForKey", "Access_ABC123").Return(mockHistoryIterator, nil)
		mockHistoryIterator.On("HasNext").Return(false)
		mockHistoryIterator.On("Close").Return(nil)

		sc := SmartContract{}
		result, err := sc.GetAccessHistory(mockCtx, "ABC123")

		assert.NoError(t, err)

		// Use JSON unmarshalling to verify structure without testing exact strings
		var historyData map[string]interface{}
		err = json.Unmarshal([]byte(result), &historyData)
		assert.NoError(t, err)

		// Check counter is 0
		counter, exists := historyData["counter"]
		assert.True(t, exists)
		assert.Equal(t, float64(0), counter)

		// Check transactions array is empty
		txns, exists := historyData["txns"]
		assert.True(t, exists)
		txnsArray, ok := txns.([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 0, len(txnsArray))
	})
}
