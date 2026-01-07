package worldmock

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/TerraDharitri/drt-go-chain-core/data/dcdt"
	"github.com/TerraDharitri/drt-go-chain-core/data/vm"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"
	"github.com/TerraDharitri/drt-go-chain-scenario/worldmock/dcdtconvert"
	vmcommon "github.com/TerraDharitri/drt-go-chain-vm-common"
)

var REWA_000000_TOKEN_IDENTIFIER = []byte("REWA-000000")

// GetTokenBalance returns the DCDT balance of an account for the given token
// key (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) GetTokenBalance(address []byte, tokenIdentifier []byte, nonce uint64) (*big.Int, error) {
	account := bf.World.AcctMap.GetAccount(address)
	if check.IfNil(account) {
		return big.NewInt(0), nil
	}
	return dcdtconvert.GetTokenBalance(tokenIdentifier, nonce, account.Storage)
}

// GetTokenData gets the DCDT information related to a token from the storage of an account
// (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) GetTokenData(address []byte, tokenIdentifier []byte, nonce uint64) (*dcdt.DCDigitalToken, error) {
	account := bf.World.AcctMap.GetAccount(address)
	if check.IfNil(account) {
		return &dcdt.DCDigitalToken{
			Value: big.NewInt(0),
		}, nil
	}
	systemAccStorage := make(map[string][]byte)
	systemAcc := bf.World.AcctMap.GetAccount(vmcommon.SystemAccountAddress)
	if systemAcc != nil {
		systemAccStorage = systemAcc.Storage
	}
	return account.GetTokenData(tokenIdentifier, nonce, systemAccStorage)
}

// SetTokenData sets the DCDT information related to a token from the storage of an account
// (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) SetTokenData(address []byte, tokenIdentifier []byte, nonce uint64, tokenData *dcdt.DCDigitalToken) error {
	account := bf.World.AcctMap.GetAccount(address)
	if check.IfNil(account) {
		return nil
	}
	return account.SetTokenData(tokenIdentifier, nonce, tokenData)
}

// ConvertToBuiltinFunction converts a VM input with a populated DCDT field into a built-in function call.
func ConvertToBuiltinFunction(tx *vmcommon.ContractCallInput) *vmcommon.ContractCallInput {
	if len(tx.DCDTTransfers) == 0 {
		return tx
	}

	if len(tx.DCDTTransfers) == 1 && !bytes.Equal(tx.DCDTTransfers[0].DCDTTokenName, REWA_000000_TOKEN_IDENTIFIER) {
		return convertToDCDTTransfer(tx, tx.DCDTTransfers[0])

	}

	return convertToMultiDCDTTransfer(tx)
}

func convertToDCDTTransfer(tx *vmcommon.ContractCallInput, dcdtTransfer *vmcommon.DCDTTransfer) *vmcommon.ContractCallInput {
	dcdtTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  tx.CallerAddr,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    tx.CallType,
			GasPrice:    tx.GasPrice,
			GasProvided: tx.GasProvided,
			GasLocked:   tx.GasLocked,
		},
		RecipientAddr:     tx.RecipientAddr,
		Function:          core.BuiltInFunctionDCDTTransfer,
		AllowInitFunction: false,
	}

	if dcdtTransfer.DCDTTokenNonce > 0 {
		dcdtTransferInput.Function = core.BuiltInFunctionDCDTNFTTransfer
		dcdtTransferInput.RecipientAddr = dcdtTransferInput.CallerAddr
		nonceAsBytes := big.NewInt(0).SetUint64(dcdtTransfer.DCDTTokenNonce).Bytes()
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments,
			dcdtTransfer.DCDTTokenName, nonceAsBytes, dcdtTransfer.DCDTValue.Bytes(), tx.RecipientAddr)
	} else {
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments,
			dcdtTransfer.DCDTTokenName, dcdtTransfer.DCDTValue.Bytes())
	}

	if len(tx.Function) > 0 {
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, []byte(tx.Function))
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, tx.Arguments...)
	}

	return dcdtTransferInput
}

func convertToMultiDCDTTransfer(tx *vmcommon.ContractCallInput) *vmcommon.ContractCallInput {
	multiTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  tx.CallerAddr,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    tx.CallType,
			GasPrice:    tx.GasPrice,
			GasProvided: tx.GasProvided,
			GasLocked:   tx.GasLocked,
		},
		RecipientAddr:     tx.CallerAddr,
		Function:          core.BuiltInFunctionMultiDCDTNFTTransfer,
		AllowInitFunction: false,
	}

	multiTransferInput.Arguments = append(multiTransferInput.Arguments, tx.RecipientAddr)

	nrTransfers := len(tx.DCDTTransfers)
	nrTransfersAsBytes := big.NewInt(0).SetUint64(uint64(nrTransfers)).Bytes()
	multiTransferInput.Arguments = append(multiTransferInput.Arguments, nrTransfersAsBytes)

	for _, dcdtTransfer := range tx.DCDTTransfers {
		token := dcdtTransfer.DCDTTokenName
		nonceAsBytes := big.NewInt(0).SetUint64(dcdtTransfer.DCDTTokenNonce).Bytes()
		value := dcdtTransfer.DCDTValue

		multiTransferInput.Arguments = append(multiTransferInput.Arguments, token, nonceAsBytes, value.Bytes())
	}

	if len(tx.Function) > 0 {
		multiTransferInput.Arguments = append(multiTransferInput.Arguments, []byte(tx.Function))
		multiTransferInput.Arguments = append(multiTransferInput.Arguments, tx.Arguments...)
	}

	return multiTransferInput
}

// PerformDirectDCDTTransfer calls the real DCDTTransfer function immediately;
// only works for in-shard transfers for now, but it will be expanded to
// cross-shard.
//
// Deprecated. TODO: remove.
func (bf *BuiltinFunctionsWrapper) PerformDirectDCDTTransfer(
	sender []byte,
	receiver []byte,
	token []byte,
	nonce uint64,
	value *big.Int,
	callType vm.CallType,
	gasLimit uint64,
	gasPrice uint64,
) (*vmcommon.VMOutput, error) {
	dcdtTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    callType,
			GasPrice:    gasPrice,
			GasProvided: gasLimit,
			GasLocked:   0,
		},
		RecipientAddr:     receiver,
		Function:          core.BuiltInFunctionDCDTTransfer,
		AllowInitFunction: false,
	}

	if nonce > 0 {
		dcdtTransferInput.Function = core.BuiltInFunctionDCDTNFTTransfer
		dcdtTransferInput.RecipientAddr = dcdtTransferInput.CallerAddr
		nonceAsBytes := big.NewInt(0).SetUint64(nonce).Bytes()
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, token, nonceAsBytes, value.Bytes(), receiver)
	} else {
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, token, value.Bytes())
	}

	vmOutput, err := bf.ProcessBuiltInFunction(dcdtTransferInput)
	if err != nil {
		return nil, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf(
			"DCDTtransfer failed: retcode = %d, msg = %s",
			vmOutput.ReturnCode,
			vmOutput.ReturnMessage)
	}

	return vmOutput, nil
}

// PerformDirectMultiDCDTTransfer -
// Deprecated. TODO: remove.
func (bf *BuiltinFunctionsWrapper) PerformDirectMultiDCDTTransfer(
	sender []byte,
	receiver []byte,
	dcdtTransfers []*scenmodel.DCDTTxData,
	callType vm.CallType,
	gasLimit uint64,
	gasPrice uint64,
) (*vmcommon.VMOutput, error) {
	nrTransfers := len(dcdtTransfers)
	nrTransfersAsBytes := big.NewInt(0).SetUint64(uint64(nrTransfers)).Bytes()

	multiTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    callType,
			GasPrice:    gasPrice,
			GasProvided: gasLimit,
			GasLocked:   0,
		},
		RecipientAddr:     sender,
		Function:          core.BuiltInFunctionMultiDCDTNFTTransfer,
		AllowInitFunction: false,
	}
	multiTransferInput.Arguments = append(multiTransferInput.Arguments, receiver, nrTransfersAsBytes)

	for i := 0; i < nrTransfers; i++ {
		token := dcdtTransfers[i].TokenIdentifier.Value
		nonceAsBytes := big.NewInt(0).SetUint64(dcdtTransfers[i].Nonce.Value).Bytes()
		value := dcdtTransfers[i].Value.Value

		multiTransferInput.Arguments = append(multiTransferInput.Arguments, token, nonceAsBytes, value.Bytes())
	}

	vmOutput, err := bf.ProcessBuiltInFunction(multiTransferInput)
	if err != nil {
		return nil, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf(
			"MultiDCDTtransfer failed: retcode = %d, msg = %s",
			vmOutput.ReturnCode,
			vmOutput.ReturnMessage)
	}

	return vmOutput, nil
}
