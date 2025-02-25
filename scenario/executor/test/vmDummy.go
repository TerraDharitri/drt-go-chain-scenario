package executortest

import (
	"errors"

	scenarioexec "github.com/TerraDharitri/drt-go-chain-scenario/scenario/executor"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"
	worldmock "github.com/TerraDharitri/drt-go-chain-scenario/worldmock"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	vmcommon "github.com/TerraDharitri/drt-go-chain-vm-common"
)

var _ scenarioexec.VMInterface = (*DummyVM)(nil)
var _ scenarioexec.VMBuilder = (*DummyVMBuilder)(nil)

// DummyVM is a VM stand-in that can never be called.
// Used for tests that do not require a VM.
type DummyVM struct{}

// RunSmartContractCreate -
func (*DummyVM) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	return nil, errors.New("cannot call the DummyVM")
}

// RunSmartContractCall -
func (*DummyVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	return nil, errors.New("cannot call the DummyVM")
}

// GasScheduleChange -
func (*DummyVM) GasScheduleChange(newGasSchedule map[string]map[string]uint64) {}

// GetVersion -
func (*DummyVM) GetVersion() string {
	return ""
}

// IsInterfaceNil -
func (*DummyVM) IsInterfaceNil() bool {
	return false
}

// Close -
func (*DummyVM) Close() error {
	return nil
}

// Reset -
func (*DummyVM) Reset() {}

// SetGasTracing -
func (*DummyVM) SetGasTracing(enableGasTracing bool) {}

// GetGasTrace -
func (*DummyVM) GetGasTrace() map[string]map[string][]uint64 {
	return make(map[string]map[string][]uint64)
}

// DummyVMBuilder is the builder for a DummyVM.
// Also provides a minimal gas schedule for running the builtin functions.
// Used for tests that do not require a VM.
type DummyVMBuilder struct{}

// NewMockWorld defines how the MockWorld is initialized.
func (*DummyVMBuilder) NewMockWorld() *worldmock.MockWorld {
	return worldmock.NewMockWorld()
}

// GasScheduleMapFromScenarios converts the gas schedule name from a scenario into an actual gas map.
func (*DummyVMBuilder) GasScheduleMapFromScenarios(scenGasSchedule scenmodel.GasSchedule) (worldmock.GasScheduleMap, error) {
	gasMap := make(map[string]map[string]uint64)
	fillGasMapInternal(gasMap, 1)
	return gasMap, nil
}

// GetVMType returns the configured VM type.
func (*DummyVMBuilder) GetVMType() []byte {
	return []byte{0, 0}
}

// NewVM creates the execution VM host with references to the world mock and gas schedule.
func (*DummyVMBuilder) NewVM(world *worldmock.MockWorld, gasSchedule map[string]map[string]uint64) (scenarioexec.VMInterface, error) {
	return &DummyVM{}, nil
}

func fillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[core.BaseOperationCostString] = fillGasMapBaseOperationCosts(value)
	gasMap[core.BuiltInCostString] = fillGasMapBuiltInCosts(value)

	return gasMap
}

func fillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value
	gasMap["AoTPreparePerByte"] = value
	gasMap["GetCode"] = value
	return gasMap
}

func fillGasMapBuiltInCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["ChangeOwnerAddress"] = value
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["SaveUserName"] = value
	gasMap["SaveKeyValue"] = value
	gasMap["DCDTTransfer"] = value
	gasMap["DCDTBurn"] = value
	gasMap["DCDTLocalMint"] = value
	gasMap["DCDTLocalBurn"] = value
	gasMap["DCDTNFTCreate"] = value
	gasMap["DCDTNFTAddQuantity"] = value
	gasMap["DCDTNFTBurn"] = value
	gasMap["DCDTNFTTransfer"] = value
	gasMap["DCDTNFTChangeCreateOwner"] = value
	gasMap["DCDTNFTAddUri"] = value
	gasMap["DCDTNFTUpdateAttributes"] = value
	gasMap["DCDTNFTMultiTransfer"] = value
	gasMap["SetGuardian"] = value
	gasMap["GuardAccount"] = value
	gasMap["UnGuardAccount"] = value
	gasMap["TrieLoadPerNode"] = value
	gasMap["TrieStorePerNode"] = value
	gasMap["DCDTModifyRoyalties"] = value
	gasMap["DCDTModifyCreator"] = value
	gasMap["DCDTNFTRecreate"] = value
	gasMap["DCDTNFTUpdate"] = value
	gasMap["DCDTNFTSetNewURIs"] = value

	return gasMap
}
