package scenjsonparse

import (
	"errors"
	"fmt"

	oj "github.com/TerraDharitri/drt-go-chain-scenario/orderedjson"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"
)

func (p *Parser) parseAccountAddress(addrRaw string) (scenmodel.JSONBytesFromString, error) {
	if len(addrRaw) == 0 {
		return scenmodel.JSONBytesFromString{}, errors.New("missing account address")
	}
	addrBytes, err := p.ExprInterpreter.InterpretString(addrRaw)
	if err == nil && len(addrBytes) != 32 {
		return scenmodel.JSONBytesFromString{}, errors.New("account address is not 32 bytes in length")
	}
	return scenmodel.NewJSONBytesFromString(addrBytes, addrRaw), err
}

func (p *Parser) processAccount(acctRaw oj.OJsonObject) (*scenmodel.Account, error) {
	acctMap, isMap := acctRaw.(*oj.OJsonMap)
	if !isMap {
		return nil, errors.New("unmarshalled account object is not a map")
	}

	acct := scenmodel.Account{
		Shard:           scenmodel.JSONUint64Zero(),
		IsSmartContract: false,
		Comment:         "",
		Nonce:           scenmodel.JSONUint64Zero(),
		Balance:         scenmodel.JSONBigIntZero(),
		Username:        scenmodel.JSONBytesEmpty(),
		Storage:         nil,
		Code:            scenmodel.JSONBytesEmpty(),
		CodeMetadata:    scenmodel.JSONBytesEmpty(),
		Owner:           scenmodel.JSONBytesEmpty(),
		AsyncCallData:   "",
		DCDTData:        nil,
		Update:          false,
		DeveloperReward: scenmodel.JSONBigIntZero(),
	}

	var err error

	for _, kvp := range acctMap.OrderedKV {
		switch kvp.Key {
		case "comment":
			acct.Comment, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account comment: %w", err)
			}
		case "update":
			acct.Update, err = p.parseBool(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid update flag bool: %w", err)
			}
		case "shard":
			acct.Shard, err = p.processUint64(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid shard number: %w", err)
			}
		case "nonce":
			acct.Nonce, err = p.processUint64(kvp.Value)
			if err != nil {
				return nil, errors.New("invalid account nonce")
			}
		case "balance":
			acct.Balance, err = p.processBigInt(kvp.Value, bigIntUnsignedBytes)
			if err != nil {
				return nil, errors.New("invalid account balance")
			}
		case "dcdt":
			dcdtMap, dcdtOk := kvp.Value.(*oj.OJsonMap)
			if !dcdtOk {
				return nil, errors.New("invalid DCDT map")
			}
			for _, dcdtKvp := range dcdtMap.OrderedKV {
				tokenNameStr, err := p.ExprInterpreter.InterpretString(dcdtKvp.Key)
				if err != nil {
					return nil, fmt.Errorf("invalid dcdt token identifer: %w", err)
				}
				tokenName := scenmodel.NewJSONBytesFromString(tokenNameStr, dcdtKvp.Key)
				dcdtItem, err := p.processDCDTData(tokenName, dcdtKvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid dcdt value: %w", err)
				}
				acct.DCDTData = append(acct.DCDTData, dcdtItem)
			}
		case "username":
			acct.Username, err = p.processStringAsByteArray(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account username: %w", err)
			}
		case "storage":
			storageMap, storageOk := kvp.Value.(*oj.OJsonMap)
			if !storageOk {
				return nil, errors.New("invalid account storage")
			}
			for _, storageKvp := range storageMap.OrderedKV {
				byteKey, err := p.ExprInterpreter.InterpretString(storageKvp.Key)
				if err != nil {
					return nil, fmt.Errorf("invalid account storage key: %w", err)
				}
				byteVal, err := p.processSubTreeAsByteArray(storageKvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account storage value: %w", err)
				}
				stElem := scenmodel.StorageKeyValuePair{
					Key:   scenmodel.NewJSONBytesFromString(byteKey, storageKvp.Key),
					Value: byteVal,
				}
				acct.Storage = append(acct.Storage, &stElem)
			}
		case "code":
			acct.Code, err = p.processStringAsByteArray(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account code: %w", err)
			}
		case "codeMetadata":
			acct.CodeMetadata, err = p.processStringAsByteArray(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account codeMetadata: %w", err)
			}
		case "owner":
			acct.Owner, err = p.processStringAsByteArray(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account owner: %w", err)
			}
		case "asyncCallData":
			acct.AsyncCallData, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid asyncCallData string: %w", err)
			}
		case "developerRewards":
			acct.DeveloperReward, err = p.processBigInt(kvp.Value, bigIntUnsignedBytes)
			if err != nil {
				return nil, errors.New("invalid developerRewards")
			}
		default:
			return nil, fmt.Errorf("unknown account field: %s", kvp.Key)
		}
	}

	return &acct, nil
}

func (p *Parser) processAccountMap(acctMapRaw oj.OJsonObject) ([]*scenmodel.Account, error) {
	var accounts []*scenmodel.Account
	preMap, isPreMap := acctMapRaw.(*oj.OJsonMap)
	if !isPreMap {
		return nil, errors.New("unmarshalled account map object is not a map")
	}
	for _, acctKVP := range preMap.OrderedKV {
		acct, acctErr := p.processAccount(acctKVP.Value)
		if acctErr != nil {
			return nil, acctErr
		}
		acctAddr, hexErr := p.parseAccountAddress(acctKVP.Key)
		if hexErr != nil {
			return nil, hexErr
		}
		acct.Address = acctAddr
		accounts = append(accounts, acct)

	}
	return accounts, nil
}
