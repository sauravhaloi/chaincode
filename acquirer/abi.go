package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

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

	// Write the state to the ledger
	err = stub.PutState("IBI-CC[init]: "+time.Now().String(), []byte("starting ABI chaincode"))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke queries another chaincode and updates its own state
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var jsonResp, customer, name, amount, state string
	var response []byte
	var err error

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	chaincodeURL := args[0] //https://github.com/sauravhaloi/chaincode/issuer
	operation := args[1]
	customer = args[2]

	switch operation {
	case "GetAccountBalance":
		f := "Query"
		queryArgs := util.ToChaincodeArgs(f, "GetAccountBalance", customer)
		response, err = stub.QueryChaincode(chaincodeURL, queryArgs)
		if err != nil {
			errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
			jsonResp = "{\"Error\":\"" + errStr + "\"}"
			return nil, errors.New(jsonResp)
		}

	case "WithdrawFund":
		f := "Invoke"
		name = strings.Split(customer, ",")[0]
		amount = strings.Split(customer, ",")[1]
		queryArgs := util.ToChaincodeArgs(f, "Withdraw", name, amount)

		response, err = stub.QueryChaincode(chaincodeURL, queryArgs)
		if err != nil {
			errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", err.Error())
			jsonResp = "{\"Error\":\"" + errStr + "\"}"
			return nil, errors.New(jsonResp)
		}

	default:
		jsonResp = "{\"Error\":\"Invalid operaton requested: " + operation + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp = "{\"Response\":\"" + string(response) + "\"}"

	// transaction was successful, charge Issuer
	settlement, err := strconv.Atoi(amount)
	if err != nil {
		return nil, err
	}

	settlement = settlement + serviceCharge

	state = fmt.Sprintf("IBI owes ABI " + strconv.Itoa(settlement))

	// Write amount which IBI owes to ABI back to the ledger
	err = stub.PutState("IBI->ABI", []byte(state))
	if err != nil {
		return nil, err
	}

	fmt.Printf("Invoke chaincode successful. IBI Owes ABI %d\n", settlement)
	return []byte(state), nil
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
		return nil, errors.New(jsonResp)
	}

	fmt.Printf("Query chaincode successful. Got IBI->ABI %s\n", string(valAsbytes))
	jsonResp = "{\"IBI Owes ABI\":\"" + string(valAsbytes) + "\"}"
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
