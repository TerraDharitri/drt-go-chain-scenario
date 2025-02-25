package scenjsonparse

import (
	"errors"
	"fmt"

	oj "github.com/TerraDharitri/drt-go-chain-scenario/orderedjson"
	scenmodel "github.com/TerraDharitri/drt-go-chain-scenario/scenario/model"
)

func (p *Parser) processNewAddressMocks(namsRaw oj.OJsonObject) ([]*scenmodel.NewAddressMock, error) {
	namList, isList := namsRaw.(*oj.OJsonList)
	if !isList {
		return nil, errors.New("newAddresses list is not a list")
	}
	var namEntries []*scenmodel.NewAddressMock
	var err error
	for _, namRaw := range namList.AsList() {
		namMap, isMap := namRaw.(*oj.OJsonMap)
		if !isMap {
			return nil, errors.New("new address mock entry is not a map")
		}
		namEntry := scenmodel.NewAddressMock{}
		for _, kvp := range namMap.OrderedKV {
			switch kvp.Key {
			case "creatorAddress":
				caStr, err := p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("creatorAddress is not a json string: %w", err)
				}
				namEntry.CreatorAddress, err = p.parseAccountAddress(caStr)
				if err != nil {
					return nil, err
				}
			case "creatorNonce":
				namEntry.CreatorNonce, err = p.processUint64(kvp.Value)
				if err != nil {
					return nil, errors.New("invalid creatorNonce")
				}
			case "newAddress":
				naStr, err := p.parseString(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("newAddress is not a json string: %w", err)
				}
				namEntry.NewAddress, err = p.parseAccountAddress(naStr)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unknown nam field: %s", kvp.Key)
			}
		}
		namEntries = append(namEntries, &namEntry)
	}

	return namEntries, nil
}
