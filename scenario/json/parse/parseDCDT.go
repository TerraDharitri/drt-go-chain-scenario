package scenjsonparse

import (
	"errors"
	"fmt"

	oj "github.com/TerraDharitri/drt-go-chain-scenario/orderedjson"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"
)

func (p *Parser) processDCDTData(
	tokenName scenmodel.JSONBytesFromString,
	dcdtDataRaw oj.OJsonObject) (*scenmodel.DCDTData, error) {

	switch data := dcdtDataRaw.(type) {
	case *oj.OJsonString:
		// simple string representing balance "400,000,000,000"
		dcdtData := scenmodel.DCDTData{
			TokenIdentifier: tokenName,
		}
		balance, err := p.processBigInt(dcdtDataRaw, bigIntUnsignedBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid DCDT balance: %w", err)
		}
		dcdtData.Instances = []*scenmodel.DCDTInstance{
			{
				Nonce:   scenmodel.JSONUint64{Value: 0, Original: ""},
				Balance: balance,
			},
		}
		return &dcdtData, nil
	case *oj.OJsonMap:
		return p.processDCDTDataMap(tokenName, data)
	default:
		return nil, errors.New("invalid JSON object for DCDT")
	}
}

// Map containing DCDT fields, e.g.:
//
//	{
//		"instances": [ ... ],
//	 "lastNonce": "5",
//		"frozen": "true"
//	}
func (p *Parser) processDCDTDataMap(tokenName scenmodel.JSONBytesFromString, dcdtDataMap *oj.OJsonMap) (*scenmodel.DCDTData, error) {
	dcdtData := scenmodel.DCDTData{
		TokenIdentifier: tokenName,
	}
	firstInstance := &scenmodel.DCDTInstance{}
	firstInstanceLoaded := false
	var explicitInstances []*scenmodel.DCDTInstance

	for _, kvp := range dcdtDataMap.OrderedKV {
		// it is allowed to load the instance directly, fields set to the first instance
		instanceFieldLoaded, err := p.tryProcessDCDTInstanceField(kvp, firstInstance)
		if err != nil {
			return nil, fmt.Errorf("invalid account DCDT instance field: %w", err)
		}
		if instanceFieldLoaded {
			firstInstanceLoaded = true
		} else {
			switch kvp.Key {
			case "instances":
				explicitInstances, err = p.processDCDTInstances(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT instances: %w", err)
				}
			case "lastNonce":
				dcdtData.LastNonce, err = p.processUint64(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT lastNonce: %w", err)
				}
			case "roles":
				dcdtData.Roles, err = p.processStringList(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT roles: %w", err)
				}
			case "frozen":
				dcdtData.Frozen, err = p.processUint64(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid DCDT frozen flag: %w", err)
				}
			default:
				return nil, fmt.Errorf("unknown DCDT data field: %s", kvp.Key)
			}
		}
	}

	if firstInstanceLoaded {
		if !p.AllowDcdtLegacySetSyntax {
			return nil, fmt.Errorf("wrong DCDT set state syntax: instances in root no longer allowed")
		}
		dcdtData.Instances = []*scenmodel.DCDTInstance{firstInstance}
	}
	dcdtData.Instances = append(dcdtData.Instances, explicitInstances...)

	return &dcdtData, nil
}

func (p *Parser) tryProcessDCDTInstanceField(kvp *oj.OJsonKeyValuePair, targetInstance *scenmodel.DCDTInstance) (bool, error) {
	var err error
	switch kvp.Key {
	case "nonce":
		targetInstance.Nonce, err = p.processUint64(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid account nonce: %w", err)
		}
	case "balance":
		targetInstance.Balance, err = p.processBigInt(kvp.Value, bigIntUnsignedBytes)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT balance: %w", err)
		}
	case "creator":
		targetInstance.Creator, err = p.processStringAsByteArray(kvp.Value)
		if err != nil || len(targetInstance.Creator.Value) != 32 {
			return false, fmt.Errorf("invalid DCDT NFT creator address: %w", err)
		}
	case "royalties":
		targetInstance.Royalties, err = p.processUint64(kvp.Value)
		if err != nil || targetInstance.Royalties.Value > 10000 {
			return false, fmt.Errorf("invalid DCDT NFT royalties: %w", err)
		}
	case "hash":
		targetInstance.Hash, err = p.processStringAsByteArray(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT hash: %w", err)
		}
	case "uri":
		targetInstance.Uris, err = p.parseValueList(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT URI: %w", err)
		}
	case "attributes":
		targetInstance.Attributes, err = p.processSubTreeAsByteArray(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT attributes: %w", err)
		}
	default:
		return false, nil
	}
	return true, nil
}

func (p *Parser) processDCDTInstances(dcdtInstancesRaw oj.OJsonObject) ([]*scenmodel.DCDTInstance, error) {
	var instancesResult []*scenmodel.DCDTInstance
	dcdtInstancesList, isList := dcdtInstancesRaw.(*oj.OJsonList)
	if !isList {
		return nil, errors.New("dcdt instances object is not a list")
	}
	for _, instanceItem := range dcdtInstancesList.AsList() {
		instanceAsMap, isMap := instanceItem.(*oj.OJsonMap)
		if !isMap {
			return nil, errors.New("JSON map expected as dcdt instances list item")
		}

		instance := &scenmodel.DCDTInstance{}

		for _, kvp := range instanceAsMap.OrderedKV {
			instanceFieldLoaded, err := p.tryProcessDCDTInstanceField(kvp, instance)
			if err != nil {
				return nil, fmt.Errorf("invalid account DCDT instance field in instances list: %w", err)
			}
			if !instanceFieldLoaded {
				return nil, fmt.Errorf("invalid account DCDT instance field in instances list: `%s`", kvp.Key)
			}
		}

		instancesResult = append(instancesResult, instance)

	}

	return instancesResult, nil
}
