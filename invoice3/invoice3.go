package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
)

//==============================================================================================================================
//	 Participant roles
//==============================================================================================================================

const   SELLER   =  "seller"
const   BUYER   =  "buyer"
const   FINANCIER =  "financier"


//==============================================================================================================================
//	Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

//==============================================================================================================================
//	Invoice - Defines the structure for a invoice object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON amount -> Struct Amount.
//==============================================================================================================================
type Invoice struct {
	InvoiceId        string `json:"invoiceid"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	Seller         string `json:"seller"`
	Buyer            string `json:"buyer"`
	DueDate          string `json:"duedate"`
	Status           string `json:"status"`
	Financier            string `json:"financier"`
	Discount         string `json:"discount"`
}


//==============================================================================================================================
//	Invoice Holder - Defines the structure that holds all the invoiceIDs for invoices that have been created.
//				     Used as an index when querying all invoices.
//==============================================================================================================================

type Invoice_Holder struct {
	Invoices 	[]string `json:"invoices"`
}


//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {


	var invoiceIDs Invoice_Holder

	bytes, err := json.Marshal(invoiceIDs)

    if err != nil { return nil, errors.New("Error creating Invoice_Holder record") }

	err = stub.PutState("invoiceIDs", bytes)
	if err != nil { return nil, errors.New("Error putting state with invoiceIDs") }

	return nil, nil
}

//==============================================================================================================================
//	 General Functions: get_username & get_role
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

	role, err := stub.ReadCertAttribute("username");
	if err != nil { return "", errors.New("Couldn't retrieve username for caller.") }
	return string(role), nil
}

func (t *SimpleChaincode) get_role(stub shim.ChaincodeStubInterface) (string, error) {

	role, err := stub.ReadCertAttribute("role");
	if err != nil { return "", errors.New("Couldn't retrieve role for caller.") }
	return string(role), nil
}


//==============================================================================================================================
//	 retrieve_invoice
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_invoice(stub shim.ChaincodeStubInterface, invoiceId string) (Invoice, error) {

	var inv Invoice

	bytes, err := stub.GetState(invoiceId);

	if err != nil { return inv, errors.New("RETRIEVE_INVOICE: Error retrieving invoice with invoice Id = " + invoiceId) }

	err = json.Unmarshal(bytes, &inv);

    if err != nil { return inv, errors.New("RETRIEVE_INVOICE: Corrupt invoice record " + string(bytes))	}

	return inv, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Vehicle struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, inv Invoice) (bool, error) {

	bytes, err := json.Marshal(inv)

	if err != nil { return false, errors.New("Error converting invoice record") }

	err = stub.PutState(inv.InvoiceId, bytes)

	if err != nil { return false, errors.New("Error storing invoice record") }

	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function.
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {


	if function == "create_invoice" {
        return t.create_invoice(stub, args)
	} else if function == "approve_trade"{
		return t.approve_trade(stub, args)
	} else if function == "reject_trade"{
		return t.reject_trade(stub, args)
	} else if function == "accept_trade"{
		return t.accept_trade(stub, args)
	}

    return nil, errors.New("Received unknown function invocation: " + function)
}
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "get_invoice_details" {
		if len(args) != 2 { return nil, errors.New("QUERY: Incorrect number of arguments passed") }
		inv, err := t.retrieve_invoice(stub, args[0])
		if err != nil { return nil, errors.New("QUERY: Error retrieving invoice "+err.Error()) }
		return t.get_invoice_details(stub, inv, args[1])
	}  else if function == "get_invoices" {
		return t.get_invoices(stub, args)
	}  else if function == "get_opening_trade_invoices" {
		return t.get_opening_trade_invoices(stub, args)
	}  else if function == "read" {											
		return t.read(stub, args)
	}  else if function == "get_username" {			
		return stub.ReadCertAttribute("username");
	}  else if function == "get_role" {
        return stub.ReadCertAttribute("role");
    }  

	return nil, errors.New("Received unknown function query " + function)

}


func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create Invoice - Creates the initial JSON for the invoice and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_invoice(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	//Args
	//				0               1              2              3            
	//			123443232        100.00           0.05         test_user1

	var inv Invoice

	var invoiceId = args[0]

	username, err := t.get_username(stub);

	invoice_json := `{ "invoiceid": "` + invoiceId + `", "amount": "` + args[1] + `", "currency": "USD", "seller": "` + username + `", "buyer": "` + args[3] + `", "duedate": "UNDEFINED", "status": "0", "financier":"UNDEFINED", "discount":"` + args[2] + `"}`

	err = json.Unmarshal([]byte(invoice_json), &inv)							// Convert the JSON defined above into a vehicle object for go

	if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(inv.InvoiceId) 								// If not an error then a record exists so cant create a new car with this V5cID as it must be unique

	if record != nil { return nil, errors.New("Invoice already exists") }

	role, err := t.get_role(stub)

	if 	role != SELLER {
		return nil, errors.New(fmt.Sprintf("Permission Denied. create_invoice. %v !== %v", role, SELLER))
	}

	_, err  = t.save_changes(stub, inv)

	if err != nil { fmt.Printf("CREATE_INVOICE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("invoiceIDs")

	if err != nil { return nil, errors.New("Unable to get invoiceIDs") }

	var invoiceIDs Invoice_Holder

	err = json.Unmarshal(bytes, &invoiceIDs)

	if err != nil {	return nil, errors.New("Corrupt Invoice_Holder record") }

	invoiceIDs.Invoices = append(invoiceIDs.Invoices, invoiceId)

	bytes, err = json.Marshal(invoiceIDs)

	if err != nil { fmt.Print("Error creating Invoice_Holder record") }

	err = stub.PutState("invoiceIDs", bytes)

	if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}



func (t *SimpleChaincode) accept_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	//Args
	//				0           
	//			123443232        
	var inv Invoice
	username, err := t.get_username(stub);
	role, err := t.get_role(stub)
	var invoiceId = args[0]


	inv, err = t.retrieve_invoice(stub, invoiceId)

	if 	role != FINANCIER {						
		return nil, errors.New(fmt.Sprintf("Permission Denied. accept_trade. %v !== %v", role, FINANCIER))
	}

	inv.Financier = username
	inv.Status = "1"

	_, err  = t.save_changes(stub, inv)

	if err != nil { fmt.Printf("OFFER_TRADE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

func (t *SimpleChaincode) approve_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	//Args
	//				0                
	//			123443232         
	var inv Invoice
	var invoiceId = args[0]

	username, err := t.get_username(stub);

	inv, err = t.retrieve_invoice(stub, invoiceId)

	if  username != inv.Buyer {
		return nil, errors.New(fmt.Sprintf("Permission Denied. approve_trade. %v !== %v", username, inv.Buyer))
	}

	inv.Status = "2"

	_, err  = t.save_changes(stub, inv)

	if err != nil { fmt.Printf("APPROVE_TRADE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

func (t *SimpleChaincode) reject_trade(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	//Args
	//				0                 
	//			123443232         
	var inv Invoice
	var invoiceId = args[0]

	username, err := t.get_username(stub);

	inv, err = t.retrieve_invoice(stub, invoiceId)

	if  username != inv.Buyer {
		return nil, errors.New(fmt.Sprintf("Permission Denied. reject_trade. %v !== %v", username, inv.Buyer))
	}

	if inv.Status == "0" {
		return nil, errors.New(fmt.Sprintf("Permission Denied. reject_trade. This invoice hasn't been bought by a third party financier"))
	}
	if inv.Status == "2" {
		return nil, errors.New(fmt.Sprintf("Permission Denied. reject_trade. This invoice has already been approved."))
	}

	inv.Status = "0"
	inv.Financier = "UNDEFINED"

	_, err  = t.save_changes(stub, inv)

	if err != nil { fmt.Printf("REJECT_TRADE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_invoice_details
//=================================================================================================================================
func (t *SimpleChaincode) get_invoice_details(stub shim.ChaincodeStubInterface, inv Invoice, caller string) ([]byte, error) {

	bytes, err := json.Marshal(inv)

	if err != nil { return nil, errors.New("GET_INVOICE_DETAILS: Invalid invoice object") }

	if 		inv.Seller  == caller		||
			inv.Buyer	== caller	||
			inv.Financier == caller	 {
				return bytes, nil
	} else {
			return nil, errors.New("Permission Denied. get_invoice_details")
	}

}

//=================================================================================================================================
//	 get_invoices
//=================================================================================================================================

func (t *SimpleChaincode) get_invoices(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	
	bytes, err := stub.GetState("invoiceIDs")
	if err != nil { return nil, errors.New("Unable to get invoiceIDs") }

	username, err := t.get_username(stub);

	var invoiceIDs Invoice_Holder

	err = json.Unmarshal(bytes, &invoiceIDs)

	if err != nil {	return nil, errors.New("Corrupt Invoice_Holder") }

	result := "["

	var temp []byte
	var inv Invoice

	for _, invoiceId := range invoiceIDs.Invoices {

		inv, err = t.retrieve_invoice(stub, invoiceId)

		if err != nil {return nil, errors.New("Failed to retrieve Invoice")}

		temp, err = t.get_invoice_details(stub, inv, username)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

func (t *SimpleChaincode) get_opening_trade_invoices(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	bytes, err := stub.GetState("invoiceIDs")

	if err != nil { return nil, errors.New("Unable to get invoiceIDs") }

	var invoiceIDs Invoice_Holder

	err = json.Unmarshal(bytes, &invoiceIDs)

	if err != nil {	return nil, errors.New("Corrupt Invoice_Holder") }

	result := "["

	var inv Invoice

	for _, invoiceId := range invoiceIDs.Invoices {

		inv, err = t.retrieve_invoice(stub, invoiceId)
		if err != nil {return nil, errors.New("Failed to retrieve Invoice")}

		if inv.Status == "0" {
			bytes, err := json.Marshal(inv)
			if err != nil { return nil, errors.New("GET_INVOICE_DETAILS: Invalid invoice object") }
			result += string(bytes) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}
