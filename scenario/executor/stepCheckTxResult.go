package scenexec

import (
	"fmt"
	"math/big"

	oj "github.com/TerraDharitri/drt-go-chain-scenario/orderedjson"
	er "github.com/TerraDharitri/drt-go-chain-scenario/scenario/expression/reconstructor"
	scenjwrite "github.com/TerraDharitri/drt-go-chain-scenario/scenario/json/write"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"

	vmcommon "github.com/TerraDharitri/drt-go-chain-vm-common"
)

func (ae *ScenarioExecutor) checkTxResults(
	txIndex string,
	blResult *scenmodel.TransactionResult,
	checkGas bool,
	output *vmcommon.VMOutput,
) error {

	if !blResult.Status.Check(big.NewInt(int64(output.ReturnCode))) {
		return fmt.Errorf("result code mismatch. Tx '%s'. Want: %s. Have: %d (%s). Message: %s",
			txIndex, blResult.Status.Original, int(output.ReturnCode), output.ReturnCode.String(), output.ReturnMessage)
	}

	if !blResult.Message.Check([]byte(output.ReturnMessage)) {
		return fmt.Errorf("result message mismatch. Tx '%s'. Want: %s. Have: %s",
			txIndex, blResult.Message.Original, output.ReturnMessage)
	}

	// check result
	if !blResult.Out.CheckList(output.ReturnData) {
		return fmt.Errorf("result mismatch. Tx '%s'. Want: %s. Have: %s",
			txIndex,
			checkBytesListPretty(blResult.Out),
			ae.exprReconstructor.ReconstructList(output.ReturnData, er.NoHint))
	}

	// check refund
	if !blResult.Refund.Check(output.GasRefund) {
		return fmt.Errorf("result gas refund mismatch. Tx '%s'. Want: %s. Have: 0x%x",
			txIndex, blResult.Refund.Original, output.GasRefund)
	}

	// check gas
	// unlike other checks, if unspecified the remaining gas check is ignored
	if checkGas && !blResult.Gas.IsUnspecified() && !blResult.Gas.Check(output.GasRemaining) {
		return fmt.Errorf("result gas mismatch. Tx '%s'. Want: %s. Got: %d (0x%x)",
			txIndex,
			blResult.Gas.Original,
			output.GasRemaining,
			output.GasRemaining)
	}

	return ae.checkTxLogs(txIndex, blResult.Logs, output.Logs)
}

func (ae *ScenarioExecutor) checkTxLogs(
	txIndex string,
	expectedLogs scenmodel.LogList,
	actualLogs []*vmcommon.LogEntry,
) error {
	// "logs": "*" means any value is accepted, log check ignored
	if expectedLogs.IsStar {
		return nil
	}

	// this is the real log check
	if len(actualLogs) < len(expectedLogs.List) {
		return fmt.Errorf("too few logs. Tx '%s'. Want:%d. Got:%d",
			txIndex,
			len(expectedLogs.List),
			len(actualLogs))
	}

	for i, actualLog := range actualLogs {
		if i < len(expectedLogs.List) {
			testLog := expectedLogs.List[i]
			err := ae.checkTxLog(txIndex, i, testLog, actualLog)
			if err != nil {
				return err
			}
		} else if !expectedLogs.MoreAllowedAtEnd {
			return fmt.Errorf("unexpected log. Tx '%s'. Log index: %d. Log:\n%s",
				txIndex,
				i,
				scenjwrite.LogToString(ae.convertLogToTestFormat(actualLog)),
			)
		}
	}

	return nil
}

func (ae *ScenarioExecutor) checkTxLog(
	txIndex string,
	logIndex int,
	expectedLog *scenmodel.LogEntry,
	actualLog *vmcommon.LogEntry) error {
	if !expectedLog.Address.Check(actualLog.Address) {
		return fmt.Errorf("bad log address. Tx '%s'. Log index: %d. Want:\n%s\nGot:\n%s",
			txIndex,
			logIndex,
			scenjwrite.LogToString(expectedLog),
			scenjwrite.LogToString(ae.convertLogToTestFormat(actualLog)))
	}
	if !expectedLog.Endpoint.Check(actualLog.Identifier) {
		return fmt.Errorf("bad log identifier. Tx '%s'. Log index: %d. Want:\n%s\nGot:\n%s",
			txIndex,
			logIndex,
			scenjwrite.LogToString(expectedLog),
			scenjwrite.LogToString(ae.convertLogToTestFormat(actualLog)))
	}
	if !expectedLog.Topics.CheckList(actualLog.Topics) {
		return fmt.Errorf("bad log topics. Tx '%s'. Log index: %d. Want: %s. Have: %s",
			txIndex,
			logIndex,
			checkBytesListPretty(expectedLog.Topics),
			ae.exprReconstructor.ReconstructList(actualLog.Topics, er.NoHint))
	}
	if !expectedLog.Data.CheckList(actualLog.Data) {
		return fmt.Errorf("bad log data. Tx '%s'. Log index: %d. Want:\n%s\nGot:\n%s",
			txIndex,
			logIndex,
			scenjwrite.LogToString(expectedLog),
			scenjwrite.LogToString(ae.convertLogToTestFormat(actualLog)))
	}
	return nil
}

// JSONCheckBytesString formats a list of JSONCheckBytes for printing to console.
// TODO: move somewhere else
func checkBytesListPretty(jcbl scenmodel.JSONCheckValueList) string {
	str := "["
	for i, jcb := range jcbl.Values {
		if i > 0 {
			str += ", "
		}

		str += oj.JSONString(jcb.Original)
	}
	return str + "]"
}

// this is a small hack, so we can reuse JSON printing in error messages
func (ae *ScenarioExecutor) convertLogToTestFormat(outputLog *vmcommon.LogEntry) *scenmodel.LogEntry {
	topics := scenmodel.JSONCheckValueList{
		Values: make([]scenmodel.JSONCheckBytes, len(outputLog.Topics)),
	}
	for i, topic := range outputLog.Topics {
		topics.Values[i] = scenmodel.JSONCheckBytesReconstructed(
			topic,
			ae.exprReconstructor.Reconstruct(topic,
				er.NoHint))
	}

	dataField := scenmodel.JSONCheckValueList{
		Values: make([]scenmodel.JSONCheckBytes, len(outputLog.Data)),
	}
	for i, data := range outputLog.Data {
		dataField.Values[i] = scenmodel.JSONCheckBytesReconstructed(
			data,
			ae.exprReconstructor.Reconstruct(data,
				er.NoHint))
	}
	testLog := scenmodel.LogEntry{
		Address: scenmodel.JSONCheckBytesReconstructed(
			outputLog.Address,
			ae.exprReconstructor.Reconstruct(outputLog.Address,
				er.AddressHint)),
		Endpoint: scenmodel.JSONCheckBytesReconstructed(
			outputLog.Identifier,
			ae.exprReconstructor.Reconstruct(outputLog.Identifier,
				er.StrHint)),
		Data:   dataField,
		Topics: topics,
	}

	return &testLog
}
