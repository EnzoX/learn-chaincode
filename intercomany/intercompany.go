package main

import (
	"fmt"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//==============================================================================================================================
//	Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}		

//==============================================================================================================================
//	Account - Defines the structure for an account object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON currency -> Struct Currency
//==============================================================================================================================
type Account struct{
	AccountNo string `json:"accountNo"`	
	DueTo string `json:"dueTo"`
	DueFrom string `json:"dueFrom"`
	Currency string `json:"currency"`				
	Period string `json:"period"`
	OpeningBalance string `json:"openingBalance"`
	Activity string `json:"activity"`
	PeriodToDateBalance string `json:"periodToDateBalance"`
	TransactionType string `json:"transactionType"`
}

var accountIndexStr = "_accountindex"	  // Define an index varibale to track all the accounts stored in the world state

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
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	_, args := stub.GetFunctionAndParameters()

	var Aval int
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting a single integer")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Expecting an integer argument to Init() for instantiate")
	}

	// Write the state to the ledger, test the network
	err = stub.PutState("test_key", []byte(strconv.Itoa(Aval)))	
	if err != nil {
		return shim.Error(err.Error())
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the account index
	err = stub.PutState(accountIndexStr, jsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	return shim.Success(nil)
}

// ============================================================================================================================
// Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		    initial arguments passed to other things for use in the called function.
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	// Handle different functions
	if function == "init" {					   //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "delete" {									
		return t.delete(stub, args)	
	} else if function == "read" {             //generic read ledger
		return t.read(stub, args)											
	} else if function == "write" {									
		return t.write(stub, args)
	} else if function == "create_account" {									
		return t.create_account(stub, args)
	} else if function == "transaction_activity" {									
		return t.transaction_activity(stub, args)										
	} else if function == "next_period" {									
		return t.next_period(stub, args)										
	}

	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

// ============================================================================================================================
// Query - legacy function
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}

// ============================================================================================================================
// Read - read a variable from chaincode world state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)	
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)										
}

// ============================================================================================================================
// Delete - remove a key/value pair from the world state
// ============================================================================================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	
	name := args[0]
	err := stub.DelState(name)													//remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	//get the account index
	accountsAsBytes, err := stub.GetState(accountIndexStr)
	if err != nil {
		return shim.Error("Failed to get account index")
	}
	var accountIndex []string
	json.Unmarshal(accountsAsBytes, &accountIndex)						
	
	//remove account from index
	for i,val := range accountIndex{
		if val == name{															//find the correct account
			accountIndex = append(accountIndex[:i], accountIndex[i+1:]...)			//remove it
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(accountIndex)									//save the new index
	err = stub.PutState(accountIndexStr, jsonAsBytes)
	return shim.Success(nil)
}

// ============================================================================================================================
// Write - directly write a variable into chaincode world state
// ============================================================================================================================
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, value string 
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]														
	value = args[1]
	err = stub.PutState(name, []byte(value))					
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// ============================================================================================================================
// Init account - create a new account, store into chaincode world state, and then append the account index
// ============================================================================================================================
func (t *SimpleChaincode) create_account(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//       0         1          2       3        4          5          6              7
	// "accountNo", "DueTo", "DueFrom", "USD", "Monthly", "45000.00", "3000.00", "Cash Transactions"

	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init acount")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}

	accountNo := args[0]

	dueTo := args[1]

	dueFrom := args[2]

	currency := args[3]

	period := args[4]

	transactionType := args[7]

	openingBalance, err := strconv.ParseFloat(args[5],64)
	if err != nil {
		return shim.Error("5th argument must be a numeric string")
	}

	activity, err := strconv.ParseFloat(args[6],64)
	if err != nil {
		return shim.Error("6th argument must be a numeric string")
	}

	periodToDateBalance := openingBalance + activity

	//check if account already exists
	accountAsBytes, err := stub.GetState(accountNo)
	if err != nil {
		return shim.Error("Failed to get account number")
	}
	res := Account{}
	json.Unmarshal(accountAsBytes, &res)
	if res.AccountNo == accountNo{
		return shim.Error("This account arleady exists")			
	}
	openingBalanceStr := strconv.FormatFloat(openingBalance, 'E', -1, 64)
	activityStr := strconv.FormatFloat(activity, 'E', -1, 64)
	periodToDateBalanceStr := strconv.FormatFloat(periodToDateBalance, 'E', -1, 64)

	//build the account json string 
	str := `{"accountno": "` + accountNo + `", "dueTo": "` + dueTo + `", "dueFrom": "` + dueFrom + `", "currency": "` + currency + `", "period": "` + period + `", "openingBalance": "` + openingBalanceStr + `", "activity": "` + activityStr + `", "periodToDateBalance": "` + periodToDateBalanceStr + `", "transactionType": "` + transactionType + `"}`
	err = stub.PutState(accountNo, []byte(str))							
	if err != nil {
		return shim.Error(err.Error())
	}
		
	//get the account index
	accountsAsBytes, err := stub.GetState(accountIndexStr)
	if err != nil {
		return shim.Error("Failed to get account index")
	}
	var accountIndex []string
	json.Unmarshal(accountsAsBytes, &accountIndex)							
	
	//append the index 
	accountIndex = append(accountIndex, accountNo)	
	jsonAsBytes, _ := json.Marshal(accountIndex)
	err = stub.PutState(accountIndexStr, jsonAsBytes)						

	return shim.Success(nil)
}

// ============================================================================================================================
// Transaction Activity - Create a transaction and change the activity balance and period-to-date balance
// ============================================================================================================================
func (t *SimpleChaincode) transaction_activity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	//      0           1  
	// "accountNo", "100.00"

	var err error
	var newActivity, newPeriodToDateBalance float64

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	amount,err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("2nd argument must be a numeric string")
	}

	account, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the first account")
	}
	res := Account{}
	json.Unmarshal(account, &res)																		
	
	Activity,err := strconv.ParseFloat(res.Activity, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	PeriodToDateBalance,err := strconv.ParseFloat(res.PeriodToDateBalance, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	newActivity = Activity + amount
	newPeriodToDateBalance = PeriodToDateBalance + amount

	newActivityStr := strconv.FormatFloat(newActivity, 'E', -1, 64)
	newPeriodToDateBalanceStr := strconv.FormatFloat(newPeriodToDateBalance, 'E', -1, 64)

	res.Activity = newActivityStr
	res.PeriodToDateBalance = newPeriodToDateBalanceStr

	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								
	if err != nil {
		return shim.Error(err.Error())
	}
	
	return shim.Success(nil)
}

// ============================================================================================================================
// Next Period - Set account to be in next period (move periodToDateBalance to openingBalance & set activity = 0)
// ============================================================================================================================
func (t *SimpleChaincode) next_period(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	//      0      
	// "accountNo"

	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	account, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the first account")
	}
	res := Account{}
	json.Unmarshal(account, &res)																		
	
	res.OpeningBalance = res.PeriodToDateBalance
	activity, err := strconv.ParseFloat("0",64)
	res.Activity = strconv.FormatFloat(activity, 'E', -1, 64)

	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								
	if err != nil {
		return shim.Error(err.Error())
	}
	
	return shim.Success(nil)
}