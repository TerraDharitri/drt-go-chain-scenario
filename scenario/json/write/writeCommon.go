package scenjsonwrite

import (
	"encoding/hex"

	oj "github.com/TerraDharitri/drt-go-chain-scenario/orderedjson"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"
)

func resultToOJ(res *scenmodel.TransactionResult) oj.OJsonObject {
	resultOJ := oj.NewMap()

	resultOJ.Put("out", checkValueListToOJ(res.Out))

	if !res.Status.IsUnspecified() {
		resultOJ.Put("status", checkBigIntToOJ(res.Status))
	}
	if !res.Message.IsUnspecified() {
		resultOJ.Put("message", checkBytesToOJ(res.Message))
	}
	if !res.Logs.IsUnspecified {
		if res.Logs.IsStar {
			resultOJ.Put("logs", stringToOJ("*"))
		} else {
			resultOJ.Put("logs", logsToOJ(res.Logs))

		}
	}
	if !res.Gas.IsUnspecified() {
		resultOJ.Put("gas", checkUint64ToOJ(res.Gas))
	}
	if !res.Refund.IsUnspecified() {
		resultOJ.Put("refund", checkBigIntToOJ(res.Refund))
	}

	return resultOJ
}

// LogToString returns a json representation of a log entry, we use it for debugging
func LogToString(logEntry *scenmodel.LogEntry) string {
	logOJ := logToOJ(logEntry)
	return oj.JSONString(logOJ)
}

func logToOJ(logEntry *scenmodel.LogEntry) oj.OJsonObject {
	logOJ := oj.NewMap()
	logOJ.Put("address", checkBytesToOJ(logEntry.Address))
	logOJ.Put("endpoint", checkBytesToOJ(logEntry.Endpoint))
	logOJ.Put("topics", checkValueListToOJ(logEntry.Topics))
	logOJ.Put("data", checkValueListToOJ(logEntry.Data))

	return logOJ
}

func logsToOJ(logEntries scenmodel.LogList) oj.OJsonObject {
	var logList []oj.OJsonObject
	for _, logEntry := range logEntries.List {
		logOJ := logToOJ(logEntry)
		logList = append(logList, logOJ)
	}
	if logEntries.MoreAllowedAtEnd {
		logList = append(logList, stringToOJ("+"))
	}
	logOJList := oj.OJsonList(logList)
	return &logOJList
}

func bigIntToOJ(i scenmodel.JSONBigInt) oj.OJsonObject {
	return &oj.OJsonString{Value: i.Original}
}

func checkBigIntToOJ(i scenmodel.JSONCheckBigInt) oj.OJsonObject {
	return &oj.OJsonString{Value: i.Original}
}

func bytesFromStringToString(bytes scenmodel.JSONBytesFromString) string {
	if len(bytes.Original) == 0 && len(bytes.Value) > 0 {
		bytes.Original = hex.EncodeToString(bytes.Value)
	}
	return bytes.Original
}

func bytesFromStringToOJ(bytes scenmodel.JSONBytesFromString) oj.OJsonObject {
	return &oj.OJsonString{Value: bytesFromStringToString(bytes)}
}

func bytesFromTreeToOJ(bytes scenmodel.JSONBytesFromTree) oj.OJsonObject {
	if bytes.OriginalEmpty() {
		bytes.Original = &oj.OJsonString{Value: hex.EncodeToString(bytes.Value)}
	}
	return bytes.Original
}

func checkBytesToOJ(checkBytes scenmodel.JSONCheckBytes) oj.OJsonObject {
	if checkBytes.OriginalEmpty() && len(checkBytes.Value) > 0 {
		checkBytes.Original = &oj.OJsonString{Value: hex.EncodeToString(checkBytes.Value)}
	}
	return checkBytes.Original
}

func valueListToOJ(jsonBytesList scenmodel.JSONValueList) oj.OJsonObject {
	var valuesList []oj.OJsonObject
	for _, blh := range jsonBytesList.Values {
		valuesList = append(valuesList, bytesFromStringToOJ(blh))
	}
	ojList := oj.OJsonList(valuesList)
	return &ojList
}

func checkValueListToOJ(jcbl scenmodel.JSONCheckValueList) oj.OJsonObject {
	if jcbl.IsStar {
		return &oj.OJsonString{Value: "*"}
	}

	var valuesList []oj.OJsonObject
	for _, jcb := range jcbl.Values {
		valuesList = append(valuesList, checkBytesToOJ(jcb))
	}
	ojList := oj.OJsonList(valuesList)
	return &ojList
}

func uint64ToOJ(i scenmodel.JSONUint64) oj.OJsonObject {
	return &oj.OJsonString{Value: i.Original}
}

func checkUint64ToOJ(i scenmodel.JSONCheckUint64) oj.OJsonObject {
	return &oj.OJsonString{Value: i.Original}
}

func stringToOJ(str string) oj.OJsonObject {
	return &oj.OJsonString{Value: str}
}

func boolToOJ(val bool) oj.OJsonObject {
	obj := oj.OJsonBool(val)
	return &obj
}
