package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/util"

	"errors"
	"strings"
)

//==============================================================================================================================
//	Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type SimpleChaincode struct {
}		

// ============================================================================================================================
//  Main - main - Starts up the chaincode
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init Function - Called when the user deploys the chaincode
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for testing the blockchain network")
	}

	// Write the state to the ledger, test the network
	err = stub.PutState("test_key", []byte(strconv.Itoa(Aval)))	
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		    initial arguments passed to other things for use in the called function.
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "init_account" {
		return t.init_account(stub, args)
	} else if function == "transfer_balance" {									
		return t.transfer_balance(stub, args)										
	}
	return nil, errors.New("Received unknown function invocation: " + function)
}

func (t *SimpleChaincode) init_account(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}

	//input sanitation
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return nil, errors.New("5th argument must be a non-empty string")
	}

	chaincodeId := args[0]

	accountNo := args[1]
	legalEntity := strings.ToLower(args[2])
	currency := args[3]
	amount := args[4]

	f := "init_account"
	invokeArgs := util.ToChaincodeArgs(f, accountNo, legalEntity, currency, amount)
	response, err := stub.InvokeChaincode(chaincodeId, invokeArgs)
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}
	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))
	err = stub.PutState(accountNo, []byte("success"))
	if err != nil {
		return nil, err
	}

	return nil, nil

}

func (t *SimpleChaincode) transfer_balance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}

	chaincodeId := args[0]
	accountFrom := args[1]
	accountTo := args[2]
	amount := args[3]

	f := "transfer_balance"
	invokeArgs := util.ToChaincodeArgs(f, accountFrom, accountTo, amount)
	response, err := stub.InvokeChaincode(chaincodeId, invokeArgs)
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}
	fmt.Printf("Invoke chaincode successful. Got response %s", string(response))

	return nil, nil

}
// ============================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "read" {												
		return t.read(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query " + function)
}

// ============================================================================================================================
// Read - read a variable from chaincode world state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)	
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil												
}

func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	chaincodeId := args[0]
	accountNo := args[1]

	f := "read"
	queryArgs := util.ToChaincodeArgs(f, accountNo)

	response, err := stub.QueryChaincode(chaincodeId, queryArgs)
	if err != nil {
		errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return nil, errors.New(errStr)
	}

	return response, nil												
}