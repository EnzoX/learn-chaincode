package main

import (
	"fmt"
	"strconv"
	"encoding/json"
	"time"

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
//	License - Defines the structure for a license object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON currency -> Struct Currency
//==============================================================================================================================
type License struct{
	LicenseKey string `json:"licenseKey"`
	LicensePartNo string `json:"licensePartNo"`	
	BaseEntityCode string `json:"baseEntityCode"`
	Quantity string `json:"quantity"`			
	LicensePrice string `json:"licensePrice"`
	SupportFee string `json:"supportFee"`
	LicenseStartDate string `json:"licenseStartDate"`
	LicenseEndDate string `json:"licenseEndDate"`
	SupportStartDate string `json:"supportStartDate"`
	SupportEndDate string `json:"supportEndDate"`
	Currency string `json:"currency"`
	LastSettlementDate string `json:"lastSettlementDate"`
}

//==============================================================================================================================
//	Entity - Defines the structure for an Entity object.
//==============================================================================================================================
type IntercompanyAccount struct{
	AccountKey string `json:"accountKey"`
	DueToEntityCode string `json:"dueToEntityCode"`
	DueFromEntityCode string `json:"dueFromEntityCode"`
	DueToEntityName string `json:"dueToEntityName"`
	DueFromEntityName string `json:"dueFromEntityName"`
	Currency string `json:"currency"`
	Period string `json:"period"`
	OpeningBalance string `json:"openingBalance"`
	Activity string `json:"activity"`
	PeriodToDateBalance string `json:"periodToDateBalance"`
	AccountNo string `json:"accountNo"`
	AccountName  string `json:"accountName"`
}

var LicenseIndexStr = "_licenseindex"	  // Define an index varibale to track all the licenses stored in the world state
var AccountIndexStr = "_accountindex"	  // Define an index varibale to track all the entities stored in the world state

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
	jsonAsBytes, _ := json.Marshal(empty)						//marshal an emtpy array of strings to clear the license & user index
	err = stub.PutState(LicenseIndexStr, jsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(AccountIndexStr, jsonAsBytes)
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
	} else if function == "read" {             //generic read ledger
		return t.read(stub, args)											
	} else if function == "create_account" {								
		return t.create_account(stub, args)
	} else if function == "create_license" {
		return t.create_license(stub, args)
	} else if function == "transfer_license" {			
		return t.transfer_license(stub, args)										
	} else if function == "delete_license" {
		return t.delete_license(stub, args)	
	} else if function == "settle_bill" {				
		return t.settle_bill(stub, args)										
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
// Create account - create a new intercompany account, store into chaincode world state, and then append the account index
// ============================================================================================================================
func (t *SimpleChaincode) create_account(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//          0                   1                  2                   3                 4           5
 	//   "DueToEntityCode", "DueFromEntityCode", "DueToEntityName", "DueFromEntityName", "Currency", "Period"
	//         6                7           8             9       
	//   "OpeningBalance", "Activity", "AccountNo", "AccountName"


	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}

	dueToEntityCode := args[0]
	dueFromEntityCode := args[1]
	accountNo := args[8]

	accountKey := dueToEntityCode + "_" + dueFromEntityCode + "_" + accountNo

	openingBalance, err := strconv.ParseFloat(args[6],64)
	if err != nil {
		return shim.Error("7th argument must be a numeric string")
	}

	activity, err := strconv.ParseFloat(args[7],64)
	if err != nil {
		return shim.Error("8th argument must be a numeric string")
	}

	periodToDateBalance := openingBalance + activity

	//check if account already exists
	accountAsBytes, err := stub.GetState(accountKey)
	if err != nil {
		return shim.Error("Failed to get account key")
	}
	res := IntercompanyAccount{}
	json.Unmarshal(accountAsBytes, &res)
	if res.AccountKey == accountKey{
		return shim.Error("This account arleady exists")			
	}

	openingBalanceStr := strconv.FormatFloat(openingBalance, 'E', -1, 64)
	activityStr := strconv.FormatFloat(activity, 'E', -1, 64)
	periodToDateBalanceStr := strconv.FormatFloat(periodToDateBalance, 'E', -1, 64)

	//build the account json string 
	str := `{"accountKey": "` + accountKey + `", "dueToEntityCode": "` + dueToEntityCode + `", "dueFromEntityCode": "` + dueFromEntityCode + `", "dueToEntityName": "` + args[2] + `", "dueFromEntityName": "` + args[3] + `", "currency": "` + args[4] + `", "period": "` + args[5] + `", "openingBalance": "` + openingBalanceStr + `", "activity": "` + activityStr + `", "periodToDateBalance": "` + periodToDateBalanceStr + `", "accountNo": "` + accountNo + `", "accountName": "` + args[9] + `"}`
	err = stub.PutState(accountKey, []byte(str))							
	if err != nil {
		return shim.Error(err.Error())
	}
		
	//get the account index
	accountsAsBytes, err := stub.GetState(AccountIndexStr)
	if err != nil {
		return shim.Error("Failed to get user index")
	}
	var accountIndex []string
	json.Unmarshal(accountsAsBytes, &accountIndex)							
	
	//append the index 
	accountIndex = append(accountIndex, accountKey)	
	jsonAsBytes, _ := json.Marshal(accountIndex)
	err = stub.PutState(AccountIndexStr, jsonAsBytes)						

	return shim.Success(nil)
}

// ============================================================================================================================
// Create license - create a new license, store into chaincode world state, and then append the license index
// ============================================================================================================================
func (t *SimpleChaincode) create_license(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//         0                 1               2             3              4                5
 	//   "LicensePartNo", "BaseEntityCode", "Quantity", "LicensePrice", "SupportFee", "LicenseStartDate"
	//         6                  7                   8              9              10
	//   "LicenseEndDate", "SupportStartDate", "SupportEndDate", "Currency", "LastSettlementDate"

	var err error
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11")
	}

	licenseKey := args[0] + "_" + args[1]

	quantity, err := strconv.ParseFloat(args[2],64)
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	licensePrice, err := strconv.ParseFloat(args[3],64)
	if err != nil {
		return shim.Error("4th argument must be a numeric string")
	}

	supportFee, err := strconv.ParseFloat(args[4],64)
	if err != nil {
		return shim.Error("5th argument must be a numeric string")
	}

	//check if license already exists
	licenseAsBytes, err := stub.GetState(licenseKey)
	if err != nil {
		return shim.Error("Failed to get license")
	}
	res := License{}
	json.Unmarshal(licenseAsBytes, &res)
	if res.LicenseKey == licenseKey{
		return shim.Error("This license arleady exists")			
	}

	quantityStr := strconv.FormatFloat(quantity, 'E', -1, 64)
	licensePriceStr := strconv.FormatFloat(licensePrice, 'E', -1, 64)
	supportFeeStr := strconv.FormatFloat(supportFee, 'E', -1, 64)

	//build the license json string 
	str := `{"licenseKey": "` + licenseKey + `", "licensePartNo": "` + args[0] + `", "baseEntityCode": "` + args[1] + `", "quantity": "` + quantityStr + `", "licensePrice": "` + licensePriceStr + `", "supportFee": "` + supportFeeStr + `", "licenseStartDate": "` + args[5] + `", "licenseEndDate": "` + args[6] + `", "supportStartDate": "` + args[7] + `", "supportEndDate": "` + args[8] + `", "currency": "` + args[9] + `", "LastSettlementDate": "` + args[10] + `"}`
	err = stub.PutState(licenseKey, []byte(str))							
	if err != nil {
		return shim.Error(err.Error())
	}
		
	//get the license index
	licensesAsBytes, err := stub.GetState(LicenseIndexStr)
	if err != nil {
		return shim.Error("Failed to get license index")
	}
	var licenseIndex []string
	json.Unmarshal(licensesAsBytes, &licenseIndex)							
	
	//append the index 
	licenseIndex = append(licenseIndex, licenseKey)	
	jsonAsBytes, _ := json.Marshal(licenseIndex)
	err = stub.PutState(LicenseIndexStr, jsonAsBytes)						

	return shim.Success(nil)
}

// ============================================================================================================================
// Transfer License - Create a transaction to transfer the license to other user
// ============================================================================================================================
func (t *SimpleChaincode) transfer_license(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	//      0                  1               2              3                   4                  5                   6
	// "LicenseKey",  "BaseEntityCode" ,  "Quantity", "LicenseAccountA", "LicenseAccountB", "SupportAccountA" , "SupportAccountB", 

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	licenseAAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the license")
	}
	resLicenseA := License{}
	json.Unmarshal(licenseAAsBytes, &resLicenseA)																

	licensePartNo := resLicenseA.licensePartNo
	originalQuantity,err := strconv.ParseFloat(resLicenseA.Quantity,64)

	licenseStartDate := resLicenseA.LicenseStartDate
	currentDate := time.Now().Format("01-02-2006")
	months := t.monthDiff(licenseStartDate,currentDate)
	licensePrice := strconv.ParseFloat(resLicenseA.LicensePrice,64)

	transferedQuantity, err := strconv.ParseFloat(args[2],64)

	licenseCharge := transferedQuantity * months * licensePrice / 60
	negLicenseCharge := -(licenseCharge)

	licenseChargeStr := strconv.FormatFloat(licenseCharge, 'E', -1, 64)
	negLicenseChargeStr := strconv.FormatFloat(negLicenseCharge, 'E', -1, 64)

	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	if (originalQuantity < transferedQuantity) {
		return shim.Error("No enough license to transfer")
	}

	newLicenseKey := licensePartNo + "_" + args[1]

	licenseBAsBytes, err := stub.GetState(newLicenseKey)
	if err != nil {
		return shim.Error("Failed to get license")
	}
	resLicenseB := License{}
	json.Unmarshal(licenseBAsBytes, &resLicenseB)

	if resLicenseB.LicenseKey == newLicenseKey{   // Has this license key
		args1 := [newLicenseKey, args[6]]
		t.settle_bill(stub, args1) // settle bill for the targeted license
		previousQuantity := strconv.ParseFloat(resLicenseB.Quantity,64)
		resLicenseB.Quantity = strconv.FormatFloat(previousQuantity + transferedQuantity, 'E', -1, 64)
		resLicenseB.LastSettlementDate = currentDate
		// update quantity and last settlement date
		licenseB, _ := json.Marshal(resLicenseB)
		err = stub.PutState(newLicenseKey, licenseB)								
		if err != nil {
			return shim.Error(err.Error())
		}
		args1 := [args[3], licenseChargeStr]
	    t.addActivityToAccount(stub,args1)
	    args2 := [args[4], negLicenseChargeStr]
	    t.addActivityToAccount(stub,args2)
		// bill the remaining license fee
	} else {
		args2 := [licensePartNo, args[1], args[2], resLicenseA.LicensePrice, resLicenseA.SupportFee, resLicenseA.LicenseStartDate, resLicenseA.LicenseEndDate,resLicenseA.SupportStartDate, resLicenseA.SupportEndDate,resLicenseA.Currency, currentDate]
		t.create_license(stub,args2)
		// create license for this key
		args1 := [args[3], licenseChargeStr]
	    t.addActivityToAccount(stub,args1)
	    args2 := [args[4], negLicenseChargeStr]
	    t.addActivityToAccount(stub,args2)
		// bill the remaining license fee
	}

	if (originalQuantity == transferedQuantity) {
		args3 := [args[0], args[5]]
		t.settle_bill(stub, args3)
		//settle bill for the original license
		args4 := [args[0]]
		t.delete_license(stub,args4)
		//delete this license key
	} else {
		args5 := [args[0], args[5]]
		t.settle_bill(stub, args5)
		//settle bill for the original license
		resLicenseA.Quantity = strconv.FormatFloat(originalQuantity - transferedQuantity, 'E', -1, 64)
		resLicenseA.LastSettlementDate = currentDate
		licenseA, _ := json.Marshal(resLicenseA)
		err = stub.PutState(args[0], licenseA)						
		if err != nil {
			return shim.Error(err.Error())
		}
		//update the quantity and last settlement date
	}
	
	return shim.Success(nil)
}

// ============================================================================================================================
// Utility Func monthDiff - Calculate month difference between two dates
// ============================================================================================================================

func (t *SimpleChaincode) monthDiff(string dateA, string dateB) int {
	var int res
	monthDateA := strconv.ParseInt(dateA[0,2],10,64)
	monthDateB := strconv.ParseInt(dateB[0,2],10,64)
	yearDateA := strconv.ParseInt(dateA[6,10],10,64)
	yearDateB := strconv.ParseInt(dateB[6,10],10,64)
	res = (yearDateB - yearDateA) * 12 + monthDateB - monthDateA
}

// ============================================================================================================================
// Utility Func addActivityToAccount - Add activity balance to account
// ============================================================================================================================

func (t *SimpleChaincode) addActivityToAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//      0            1
	// "accountKey", "Amount"

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	account, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the account")
	}
	resAccount := IntercompanyAccount{}
	json.Unmarshal(account, &resAccount)

	amount := strconv.ParseFloat(args[1],64)

	activity := strconv.ParseFloat(resAccount.Activity,64)
	newActivity := activity + amount
	newActivityStr := strconv.FormatFloat(newActivity, 'E', -1, 64)
	resAccount.Activity = newActivityStr

	periodToDateBalance := strconv.ParseFloat(resAccount.PeriodToDateBalance,64)
	newPeriodToDateBalance := periodToDateBalance + amount
	newPeriodToDateBalanceStr := strconv.FormatFloat(newPeriodToDateBalance, 'E', -1, 64)
	resAccount.PeriodToDateBalance = newPeriodToDateBalanceStr

	accountAsBytes, _ := json.Marshal(resAccount)
	err = stub.PutState(args[1], accountAsBytes)								
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// ============================================================================================================================
// Settle Bill - Create a transaction to settle bill for the license at the end of the period
// ============================================================================================================================
func (t *SimpleChaincode) settle_bill(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	//      0             1
	// "licenseKey", "accountKey"

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	currentDate := time.Now().Format("01-02-2006")

	license, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the license")
	}
	resLicense := License{}
	json.Unmarshal(license, &resLicense)	

	lastSettlementDate := resLicense.LastSettlementDate

	months := t.monthDiff(lastSettlementDate, currentDate)

	quantity := strconv.ParseFloat(resLicense.Quantity,64)

	supportFee := strconv.ParseFloat(resLicense.SupportFee,64)

	supportCharge := supportFee * quantity * months / 12

	supportChargeStr := strconv.FormatFloat(supportCharge, 'E', -1, 64)

	args1 := [args[1], supportChargeStr]
	t.addActivityToAccount(stub,args1)
	
	resLicense.LastSettlementDate = currentDate
	licenseAsBytes, _ := json.Marshal(resLicense)
	err = stub.PutState(args[0], licenseAsBytes)								
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}


// ============================================================================================================================
// Next Period - Roll into next period for a specific account, usually execute in the beginning of next month
// ============================================================================================================================
func (t *SimpleChaincode) next_period(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	//      0    
	// "accountKey"

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	account, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get the account")
	}
	resAccount := IntercompanyAccount{}
	json.Unmarshal(account, &resAccount)

	monthPeriod := resAccount.Period[0,3]
	yearPeriod := strconv.ParseInt(Period[4,6],10,64)

	var months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]

	for i := 0; i < len(months); i++ {
		if monthPeriod == months[i] {
			if (i < len(months) - 1 ){
				newMonthPeriod := months[i+1]
				newYearPeriod := strconv.FormatInt(yearPeriod, 10)
			} else {
				newMonthPeriod := "Jan"
				newYearPeriod := strconv.FormatInt(yearPeriod+1, 10)
			}
		}
	}

	newPeriod := newMonthPeriod + "-" + newYearPeriod

	resAccount.Period = newPeriod

	resAccount.OpeningBalance = resAccount.PeriodToDateBalance

	resAccount.Activity = strconv.FormatFloat("0", 'E', -1, 64)

	accountAsBytes, _ := json.Marshal(resAccount)
	err = stub.PutState(args[1], accountAsBytes)								
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// ============================================================================================================================
// Delete License - remove a license from the world state
// ============================================================================================================================
func (t *SimpleChaincode) delete_license(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//      0    
	// "licenseKey"
	
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	
	licenseKey := args[0]
	err := stub.DelState(licenseKey)													//remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	//get the license index
	licensesAsBytes, err := stub.GetState(LicenseIndexStr)
	if err != nil {
		return shim.Error("Failed to get license index")
	}
	var licenseIndex []string
	json.Unmarshal(licensesAsBytes, &licenseIndex)						
	
	//remove license from index
	for i,val := range licenseIndex{
		if val == licenseKey{													    //find the correct license
			licenseIndex = append(licenseIndex[:i], licenseIndex[i+1:]...)			//remove it
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(licenseIndex)									//save the new index
	err = stub.PutState(LicenseIndexStr, jsonAsBytes)
	return shim.Success(nil)
}