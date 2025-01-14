package scenjsonparse

import (
	fr "github.com/TerraDharitri/drt-go-chain-scenario/scenario/expression/fileresolver"
	ei "github.com/TerraDharitri/drt-go-chain-scenario/scenario/expression/interpreter"
)

// Parser performs parsing of both json tests (older) and scenarios (new).
type Parser struct {
	ExprInterpreter                  ei.ExprInterpreter
	AllowDcdtTxLegacySyntax          bool
	AllowDcdtLegacySetSyntax         bool
	AllowDcdtLegacyCheckSyntax       bool
	AllowSingleValueInCheckValueList bool
}

// NewParser provides a new Parser instance.
func NewParser(fileResolver fr.FileResolver, vmType []byte) Parser {
	return Parser{
		ExprInterpreter: ei.ExprInterpreter{
			FileResolver: fileResolver,
			VMType:       vmType,
		},
		AllowDcdtTxLegacySyntax:          true,
		AllowDcdtLegacySetSyntax:         true,
		AllowDcdtLegacyCheckSyntax:       true,
		AllowSingleValueInCheckValueList: true,
	}
}
