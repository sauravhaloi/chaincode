package main

import (
	"errors"
	"fmt"
	"strconv"

	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/util"
)

var serviceCharge = 5

var logger = shim.NewLogger("ftLogger")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Init initializes the chaincode
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error
	if len(args) > 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	// Initialize the chaincode
	err = stub.PutState("IBI->ABI", []byte(strconv.Itoa(0)))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return nil, nil
}

// Invoke queries another chaincode and updates its own state
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var jsonResp, customer, name, amount, state string
	var response []byte
	var err error
	var settlement, pastSettlement int

	logger.Info("abi Invoke args: ", args)

	chaincodeURL := args[0] //https://github.com/sauravhaloi/chaincode/issuer
	operation := args[1]
	customer = args[2]

	switch operation {
	case "GetAccountBalance":
		f := "GetAccountBalance"
		queryArgs := util.ToChaincodeArgs(f, customer)
		response, err = stub.QueryChaincode(chaincodeURL, queryArgs)
		if err != nil {
			errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
			jsonResp = "{\"Error\":\"" + errStr + "\"}"
			logger.Error(jsonResp)
			return nil, errors.New(jsonResp)
		}

		logger.Infof("Operation: %s | Response: %s", operation, string(response))

	case "WithdrawFund":
		f := "Withdraw"
		name = strings.Split(customer, ",")[0]
		amount = strings.Split(customer, ",")[1]
		queryArgs := util.ToChaincodeArgs(f, name, amount)

		fmt.Println("Query Args: ", queryArgs)

		response, err = stub.InvokeChaincode(chaincodeURL, queryArgs)
		if err != nil {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
			jsonResp = "{\"Error\":\"" + errStr + "\"}"
			logger.Error(jsonResp)
			return nil, errors.New(jsonResp)
		}

		logger.Infof("Operation: %s | Response: %s", operation, string(response))

		jsonResp = "{\"Response of WithdrawFund\":\"" + string(response) + "\"}"
		logger.Infof("Operation: %s | Response: %s", operation, jsonResp)

		// transaction was successful, charge Issuer
		settlement, err = strconv.Atoi(amount)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		settlement = settlement + serviceCharge

		existingAsBytes, err := stub.GetState("IBI->ABI")
		if err != nil {
			logger.Error(err)
			return nil, errors.New("Failed to get state" + err.Error())
		}

		pastSettlement, err = strconv.Atoi(string(existingAsBytes))
		if err != nil {
			logger.Error(err)
			return nil, errors.New("Error retrieving past settlement: " + err.Error())
		}

		state = strconv.Itoa(settlement + pastSettlement)
		logger.Info("state: ", state)

		// Write amount which IBI owes to ABI back to the ledger
		err = stub.PutState("IBI->ABI", []byte(state))
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		fmt.Printf("Invoke chaincode successful. IBI Owes ABI %d\n", settlement)
		return []byte(state), nil

	case "Settlement":
		logger.Info("Running Settlement")

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

		logger.Info("IBI has paid back to ABI all dues, commit it in the ledger")
		err = stub.PutState("IBI->ABI", dueAsBytes)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

	default:
		jsonResp = "{\"Error\":\"Invalid operaton requested: " + operation + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp = "{\"Response\":\"" + string(response) + "\"}"
	fmt.Printf("Operation: %s | Response: %s", operation, jsonResp)

	return []byte(jsonResp), nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function != "Query" {
		return nil, errors.New("Invalid query function name. Expecting \"Query\"")
	}
	var jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	transactionName := args[0]

	valAsbytes, err := stub.GetState(transactionName)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + transactionName + "\"}"
		logger.Error(jsonResp)
		return nil, errors.New(jsonResp)
	}

	fmt.Printf("Query chaincode successful. Got IBI->ABI %s\n", string(valAsbytes))
	jsonResp = "{\"IBI Owes ABI\":\"" + string(valAsbytes) + "\"}"
	logger.Error(jsonResp)
	fmt.Printf("Query Response:%s\n", jsonResp)
	return []byte(valAsbytes), nil

}

func main() {
	var err error
	lld, _ := shim.LogLevel("DEBUG")
	fmt.Println(lld)

	logger.SetLevel(lld)
	fmt.Println(logger.IsEnabledFor(lld))

	err = shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Could not start SimpleChaincode")
	} else {
		fmt.Println("SimpleChaincode successfully started")
	}
}
