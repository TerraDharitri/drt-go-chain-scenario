package scenjsonparse

import (
	"errors"
	"fmt"

	oj "github.com/TerraDharitri/drt-go-chain-scenario/orderedjson"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"
)

// ParseScenarioFile converts a scenario json string to scenario object representation
func (p *Parser) ParseScenarioFile(jsonString []byte) (*scenmodel.Scenario, error) {
	jobj, err := oj.ParseOrderedJSON(jsonString)
	if err != nil {
		return nil, err
	}

	topMap, isMap := jobj.(*oj.OJsonMap)
	if !isMap {
		return nil, errors.New("unmarshalled test top level object is not a map")
	}

	scenario := &scenmodel.Scenario{
		CheckGas:    true,
		TraceGas:    false,
		GasSchedule: scenmodel.GasScheduleDefault,
	}

	for _, kvp := range topMap.OrderedKV {
		switch kvp.Key {
		case "name":
			scenario.Name, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("bad scenario name: %w", err)
			}
		case "comment":
			scenario.Comment, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("bad scenario comment: %w", err)
			}
		case "checkGas":
			checkGasOJ, isBool := kvp.Value.(*oj.OJsonBool)
			if !isBool {
				return nil, errors.New("scenario checkGas flag is not boolean")
			}
			scenario.CheckGas = bool(*checkGasOJ)
		case "traceGas":
			traceGasOJ, isBool := kvp.Value.(*oj.OJsonBool)
			if !isBool {
				return nil, errors.New("scenario traceGas flag is not boolean")
			}
			scenario.TraceGas = bool(*traceGasOJ)
		case "gasSchedule":
			scenario.GasSchedule, err = p.parseGasSchedule(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("bad scenario gasSchedule: %w", err)
			}
		case "steps":
			scenario.Steps, err = p.processScenarioStepList(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("error processing steps: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown scenario field: %s", kvp.Key)
		}
	}
	return scenario, nil
}

func (p *Parser) parseGasSchedule(value oj.OJsonObject) (scenmodel.GasSchedule, error) {
	gasScheduleStr, err := p.parseString(value)
	if err != nil {
		return scenmodel.GasScheduleDummy, fmt.Errorf("gasSchedule type not a string: %w", err)
	}
	switch gasScheduleStr {
	case "default":
		return scenmodel.GasScheduleDefault, nil
	case "dummy":
		return scenmodel.GasScheduleDummy, nil
	case "v3":
		return scenmodel.GasScheduleV3, nil
	case "v4":
		return scenmodel.GasScheduleV4, nil
	default:
		return scenmodel.GasScheduleDummy, fmt.Errorf("invalid gasSchedule: %s", gasScheduleStr)
	}
}

func (p *Parser) processScenarioStepList(obj interface{}) ([]scenmodel.Step, error) {
	listRaw, listOk := obj.(*oj.OJsonList)
	if !listOk {
		return nil, errors.New("steps not a JSON list")
	}
	var stepList []scenmodel.Step
	for _, elemRaw := range listRaw.AsList() {
		step, err := p.processScenarioStep(elemRaw)
		if err != nil {
			return nil, err
		}
		stepList = append(stepList, step)
	}
	return stepList, nil
}

// ParseScenarioStep parses a single scenario step, instead of an entire file.
// Handy for tests, where step snippets can be embedded in code.
func (p *Parser) ParseScenarioStep(jsonSnippet string) (scenmodel.Step, error) {
	jobj, err := oj.ParseOrderedJSON([]byte(jsonSnippet))
	if err != nil {
		return nil, err
	}

	return p.processScenarioStep(jobj)
}

func (p *Parser) processScenarioStep(stepObj oj.OJsonObject) (scenmodel.Step, error) {
	stepMap, isStepMap := stepObj.(*oj.OJsonMap)
	if !isStepMap {
		return nil, errors.New("unmarshalled step object is not a map")
	}

	var err error
	stepTypeStr := ""
	for _, kvp := range stepMap.OrderedKV {
		if kvp.Key == "step" {
			stepTypeStr, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("step type not a string: %w", err)
			}
		}
	}

	switch stepTypeStr {
	case "":
		return nil, errors.New("no step type field provided")
	case scenmodel.StepNameExternalSteps:
		traceGasStatus := scenmodel.Undefined
		step := &scenmodel.ExternalStepsStep{TraceGas: traceGasStatus}
		for _, kvp := range stepMap.OrderedKV {
			switch kvp.Key {
			case "step":
			case "comment":
				step.Comment, err = p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("bad externalSteps step comment: %w", err)
				}
			case "traceGas":
				traceGasOJ, isBool := kvp.Value.(*oj.OJsonBool)
				if !isBool {
					return nil, errors.New("scenario traceGas flag is not boolean")
				}
				if *traceGasOJ {
					step.TraceGas = 1
				} else {
					step.TraceGas = 0
				}
			case "path":
				step.Path, err = p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("bad externalSteps path: %w", err)
				}
			default:
				return nil, fmt.Errorf("invalid externalSteps field: %s", kvp.Key)
			}
		}
		return step, nil
	case scenmodel.StepNameSetState:
		step := &scenmodel.SetStateStep{}
		for _, kvp := range stepMap.OrderedKV {
			switch kvp.Key {
			case "step":
			case "id":
				step.SetStateIdent, err = p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("bad tx set state step id: %w", err)
				}
			case "comment":
				step.Comment, err = p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("bad set state step comment: %w", err)
				}
			case "accounts":
				step.Accounts, err = p.processAccountMap(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("cannot parse set state step: %w", err)
				}
			case "newAddresses":
				step.NewAddressMocks, err = p.processNewAddressMocks(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("error parsing new addresses: %w", err)
				}
			case "previousBlockInfo":
				step.PreviousBlockInfo, err = p.processBlockInfo(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("error parsing previousBlockInfo: %w", err)
				}
			case "currentBlockInfo":
				step.CurrentBlockInfo, err = p.processBlockInfo(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("error parsing currentBlockInfo: %w", err)
				}
			case "blockHashes":
				step.BlockHashes, err = p.parseValueList(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("error parsing block hashes: %w", err)
				}
			default:
				return nil, fmt.Errorf("invalid set state field: %s", kvp.Key)
			}
		}
		return step, nil
	case scenmodel.StepNameCheckState:
		step := &scenmodel.CheckStateStep{}
		for _, kvp := range stepMap.OrderedKV {
			switch kvp.Key {
			case "step":
			case "id":
				step.CheckStateIdent, err = p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("bad check state step id: %w", err)
				}
			case "comment":
				step.Comment, err = p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("bad check state step comment: %w", err)
				}
			case "accounts":
				step.CheckAccounts, err = p.processCheckAccountMap(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("cannot parse check state step: %w", err)
				}
			default:
				return nil, fmt.Errorf("invalid check state field: %s", kvp.Key)
			}
		}
		return step, nil
	case scenmodel.StepNameDumpState:
		step := &scenmodel.DumpStateStep{}
		for _, kvp := range stepMap.OrderedKV {
			switch kvp.Key {
			case "step":
			case "comment":
				step.Comment, err = p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("bad check state step comment: %w", err)
				}
			default:
				return nil, fmt.Errorf("invalid check state field: %s", kvp.Key)
			}
		}
		return step, nil
	case scenmodel.StepNameScCall:
		return p.parseTxStep(scenmodel.ScCall, stepMap)
	case scenmodel.StepNameScDeploy:
		return p.parseTxStep(scenmodel.ScDeploy, stepMap)
	case scenmodel.StepNameScUpgrade:
		return p.parseTxStep(scenmodel.ScUpgrade, stepMap)
	case scenmodel.StepNameScQuery:
		return p.parseTxStep(scenmodel.ScQuery, stepMap)
	case scenmodel.StepNameTransfer:
		return p.parseTxStep(scenmodel.Transfer, stepMap)
	case scenmodel.StepNameValidatorReward:
		return p.parseTxStep(scenmodel.ValidatorReward, stepMap)
	default:
		return nil, fmt.Errorf("unknown step type: %s", stepTypeStr)
	}
}

func (p *Parser) parseTxStep(txType scenmodel.TransactionType, stepMap *oj.OJsonMap) (*scenmodel.TxStep, error) {
	step := &scenmodel.TxStep{}
	var err error
	for _, kvp := range stepMap.OrderedKV {
		switch kvp.Key {
		case "step":
		case "txId":
			fallthrough
		case "id":
			step.TxIdent, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("bad tx step id: %w", err)
			}
		case "displayLogs":
			step.DisplayLogs, err = p.parseBool(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("bad tx step displayLogs: %w", err)
			}
		case "comment":
			step.Comment, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("bad tx step comment: %w", err)
			}
		case "tx":
			step.Tx, err = p.processTx(txType, kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse tx step transaction: %w", err)
			}
		case "expect":
			if !step.Tx.Type.IsSmartContractTx() {
				return nil, fmt.Errorf("no expected result allowed for step of type %s", step.StepTypeName())
			}
			step.ExpectedResult, err = p.processTxExpectedResult(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse tx expected result: %w", err)
			}
		default:
			return nil, fmt.Errorf("invalid tx step field: %s", kvp.Key)
		}
	}
	return step, nil
}
