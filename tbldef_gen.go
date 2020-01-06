package main

import (
	"fmt"

	"github.com/iancoleman/strcase"
)

// case変換用の拡張メソッドを生やしたstring
type caseString string

func (cs caseString) LCamel() caseString {
	return caseString(strcase.ToLowerCamel(string(cs)))
}

func (cs caseString) UCamel() caseString {
	return caseString(strcase.ToCamel(string(cs)))
}

func (cs caseString) Quoted() caseString {
	return caseString(`"` + string(cs) + `"`)
}

// テーブル定義yamlファイル上でインデックスのキー属性のデータ型を指定する文字列
// GoType()で対応するGoの型名に変換できる
type attrType string

var attrType2GoType = map[string]string{
	"string": "string",
	"int":    "int",
	"int64":  "int64",
	"bytes":  "[]byte",
}

func (at attrType) IsValid() bool {
	_, ok := attrType2GoType[string(at)]
	return ok
}

func (at attrType) GoType() (string, error) {
	if !at.IsValid() {
		return "", fmt.Errorf("invalid attr type: %s", string(at))
	}
	return attrType2GoType[string(at)], nil
}

type tableGenDef struct {
	TblName   caseString
	PrimIdx   indexGenDef
	ScndIdxes []indexGenDef
}

type indexGenDef struct {
	TblName caseString
	IdxName caseString
	PK      attrGenDef
	SK      attrGenDef
}

func (i indexGenDef) IsSimple() bool {
	return i.SK.isEmpty()
}

type attrGenDef struct {
	AttrName caseString
	Type     attrType
}

func (a attrGenDef) isEmpty() bool {
	return a.AttrName == "" && a.Type == ""
}
