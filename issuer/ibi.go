package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/util"
)

var account = make(map[string]int)

var logger = shim.NewLogger("ftLogger")

// SampleChaincode struct required to implement the shim.Chaincode interface
type SampleChaincode struct {
}

// Init method is called when the chaincode is first deployed onto the blockchain network
func (t *SampleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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
		return nil, errors.New("Expecting integer value for customer account balance: " + err.Error())
	}

	logger.Infof("Customer: %s, Available Balance: %d", customerName, currentBalance)

	// Save the Customer info
	account[customerName] = currentBalance

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

	logger.Info("ibi Invoke args: ", args)

	// Handle different functions
	switch function {
	case "Init":
		return t.Init(stub, "init", args)
	case "Deposit":
		return t.depositFund(stub, args)
	case "Withdraw":
		return t.withdrawFund(stub, args)
	case "Settlement":
		return t.eodSettlement(stub, args)
	default:
		logger.Error("Invoke did not find func: " + function)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *SampleChaincode) getAccountBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("Running getAccountBalance")
	var name string

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]

	valAsbytes := []byte(strconv.Itoa(account[name]))

	return valAsbytes, nil
}

func (t *SampleChaincode) depositFund(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("Running depositFund")

	var name, jsonResp string
	var value, currentBalance, newBalance int
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the customer and value to set")
	}

	name = args[0]
	value, err = strconv.Atoi(args[1])
	if err != nil {
		jsonResp = "{\"Error\":\"Invalid input argument for deposit amount: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	currentBalance = account[name]
	newBalance = value + currentBalance

	account[name] = newBalance

	valAsbytes := []byte(strconv.Itoa(account[name]))

	return valAsbytes, nil
}

func (t *SampleChaincode) withdrawFund(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("Running withdrawFund")

	var name, jsonResp string
	var value, currentBalance, newBalance int
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the customer and value to set")
	}

	name = args[0]
	value, err = strconv.Atoi(args[1])
	if err != nil {
		jsonResp = "{\"Error\":\"Invalid input argument for withdraw amount: " + err.Error() + "\"}"
		return nil, errors.New(jsonResp)
	}

	currentBalance = account[name]

	if value > currentBalance {
		jsonResp = "{\"Error\":\"Insufficient Fund in account. Aborting...\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	newBalance = currentBalance - value

	account[name] = newBalance

	valAsbytes := []byte(strconv.Itoa(account[name]))

	return valAsbytes, nil
}

func (t *SampleChaincode) eodSettlement(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("Running eodSettlement")

	var err error
	var response []byte
	var jsonResp string

	logger.Info("ibi eodSettlement args: ", args)

	chaincodeURL := args[0] //https://github.com/sauravhaloi/chaincode/issuer
	operation := args[1]
	customer := args[2]

	queryArgs := util.ToChaincodeArgs(operation, chaincodeURL, operation, customer)

	response, err = stub.InvokeChaincode(chaincodeURL, queryArgs)
	if err != nil {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
		jsonResp = "{\"Error\":\"" + errStr + "\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	logger.Infof("Operation: %s | Response: %s", operation, string(response))

	/*

		dueAsBytes, err := stub.GetState("IBI->ABI")
		if err != nil {
			logger.Error(err)
			return nil, errors.New("Failed to get state" + err.Error())
		}

		if dueAsBytes == nil {
			logger.Info("IBI does not owe any dues to ABI")
			return nil, nil
		}

		logger.Infof("IBI owes %s to ABI", string(dueAsBytes))

		logger.Info("IBI pays back to ABI all dues, commit it in the ledger")
		err = stub.PutState("IBI->ABI", dueAsBytes)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

	*/

	logger.Info("EOD Settlement Done!")

	return nil, err
}

func main() {
	var err error
	lld, _ := shim.LogLevel("DEBUG")
	fmt.Println(lld)

	logger.SetLevel(lld)
	fmt.Println(logger.IsEnabledFor(lld))

	err = shim.Start(new(SampleChaincode))
	if err != nil {
		fmt.Println("Could not start SampleChaincode")
	} else {
		fmt.Println("SampleChaincode successfully started")
	}
}
