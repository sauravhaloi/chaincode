package main

import (
	"errors"
	"fmt"
	"strconv"

	"encoding/json"

	"io/ioutil"

	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("ftLogger")

// cbsDB is a JSON file which emulates a private own database
var cbsDB = "/tmp/ibi_cbs.json"

// SampleChaincode struct required to implement the shim.Chaincode interface
type SampleChaincode struct {
}

// Init method is called when the chaincode is first deployed onto the blockchain network
func (t *SampleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) > 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	var customerName string // Name of the customer
	var currentBalance int  // Current account balance of the customer
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	// Initialize the chaincode
	customerName = args[0]
	currentBalance, err = strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("Expecting integer value for customer account balance")
	}

	logger.Info("Customer: %s Available Balance: %d", customerName, currentBalance)

	// Save the Customer info
	db := make(map[string]interface{})
	db["cName"] = customerName
	db["cBalance"] = currentBalance

	jsonDB, err := json.Marshal(db)
	if err != nil {
		logger.Error("Unable to initialize core private database")
		return nil, err
	}

	ioutil.WriteFile(cbsDB, jsonDB, 0777)

	// Write the state to the ledger
	err = stub.PutState("IBI-CC[init]: "+time.Now().String(), []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Query method is invoked whenever any read/get/query operation needs to be performed on the blockchain state.
func (t *SampleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Info("Query is running: " + function)

	// Handle different functions
	if function == "GetAccountBalance" { //read a variable
		return t.getAccountBalance(stub, args)
	}
	logger.Error("Query did not find func: " + function)

	return nil, errors.New("Received unknown function query")
}

// Invoke method is invoked whenever the state of the blockchain is to be modified.
func (t *SampleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Info("Invoke is running: " + function)

	// Handle different functions
	switch function {
	case "Init":
		return t.Init(stub, "init", args)
	case "Deposit":
		return t.depositFund(stub, args)
	case "Withdraw":
		return t.withdrawFund(stub, args)
	default:
		logger.Error("Invoke did not find func: " + function)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *SampleChaincode) getAccountBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("Running getAccountBalance")
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := ioutil.ReadFile(cbsDB)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get account balance for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

func (t *SampleChaincode) depositFund(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("Running depositFund")

	var name, jsonResp string
	var value, currentBalance, newBalance int
	var err error
	var db = make(map[string]interface{})

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the customer and value to set")
	}

	name = args[0]
	value, err = strconv.Atoi(args[1])
	if err != nil {
		jsonResp = "{\"Error\":\"Invalid input argument for deposit amount: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	valAsBytes, err := ioutil.ReadFile(cbsDB)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to connect to DB: + " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	err = json.Unmarshal(valAsBytes, db)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to retrieve account details: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	if db["cName"].(string) != name {
		jsonResp = "{\"Error\":\"Invalid customer name!\"}"
		return nil, errors.New(jsonResp)
	}

	currentBalance, err = strconv.Atoi(db["cBalance"].(string))
	if err != nil {
		jsonResp = "{\"Error\":\"Internal error in fund deposit: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	newBalance = value + currentBalance

	db["cBalance"] = strconv.Itoa(newBalance)

	jsonDB, err := json.Marshal(db)
	if err != nil {
		jsonResp = "{\"Error\":\"Error in fund deposit: " + err.Error() + "\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	err = ioutil.WriteFile(cbsDB, jsonDB, 0777)
	if err != nil {
		jsonResp = "{\"Error\":\"Error in DB commit: " + err.Error() + "\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	return nil, nil
}

func (t *SampleChaincode) withdrawFund(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("Running withdrawFund")

	var name, jsonResp string
	var value, currentBalance, newBalance int
	var err error
	var db = make(map[string]interface{})

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the customer and value to set")
	}

	name = args[0]
	value, err = strconv.Atoi(args[1])
	if err != nil {
		jsonResp = "{\"Error\":\"Invalid input argument for withdraw amount: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	valAsBytes, err := ioutil.ReadFile(cbsDB)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to connect to DB: + " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	err = json.Unmarshal(valAsBytes, db)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to retrieve account details: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	if db["cName"].(string) != name {
		jsonResp = "{\"Error\":\"Invalid customer name!\"}"
		return nil, errors.New(jsonResp)
	}

	currentBalance, err = strconv.Atoi(db["cBalance"].(string))
	if err != nil {
		jsonResp = "{\"Error\":\"Internal error in fund withdraw: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	if value > currentBalance {
		jsonResp = "{\"Error\":\"Insufficient Fund in account. Aborting...\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	newBalance = value + currentBalance

	db["cBalance"] = strconv.Itoa(newBalance)

	jsonDB, err := json.Marshal(db)
	if err != nil {
		jsonResp = "{\"Error\":\"Error in fund withdraw: " + err.Error() + "\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	err = ioutil.WriteFile(cbsDB, jsonDB, 0777)
	if err != nil {
		jsonResp = "{\"Error\":\"Error in DB commit: " + err.Error() + "\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	return nil, nil
}

func main() {
	lld, _ := shim.LogLevel("DEBUG")
	fmt.Println(lld)

	logger.SetLevel(lld)
	fmt.Println(logger.IsEnabledFor(lld))
	err := shim.Start(new(SampleChaincode))
	if err != nil {
		fmt.Println("Could not start SampleChaincode")
	} else {
		fmt.Println("SampleChaincode successfully started")
	}
}
