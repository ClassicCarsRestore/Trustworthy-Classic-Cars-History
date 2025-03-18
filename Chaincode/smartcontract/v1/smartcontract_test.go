package v1

import (
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"classicschain/chaincode/mocks/github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSmartContract_InitLedger(t *testing.T) {
	mockCtx := new(contractapi.MockTransactionContextInterface)
	mockStub := new(shim.MockChaincodeStubInterface)
	mockCtx.On("GetStub").Return(mockStub)

	mockStub.On("PutState", "Classic_07642", mock.Anything).Return(nil)
	mockStub.On("GetCreator").Return([]byte("test-creator"), nil)

	sc := SmartContract{}
	err := sc.InitLedger(mockCtx)

	assert.NoError(t, err)
	mockStub.AssertCalled(t, "PutState", "Classic_07642", mock.Anything)
	mockStub.AssertCalled(t, "GetCreator")
}
