package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
)

func generateAll(defFilename string, outDir string) {
	defs := make([]tableDef, 0)
	defFile, err := os.Open(defFilename)
	if err != nil {
		log.Println(err)
		return
	}
	err = yaml.NewDecoder(defFile).Decode(&defs)
	if err != nil {
		log.Println(err)
		return
	}

	for _, tblDef := range defs {
		if !tblDef.isValid() {
			log.Printf("invalid table definition. table name: %s", tblDef.Name)
			continue
		}

		filename := fmt.Sprintf("%s.go", strcase.ToSnake(tblDef.Name))
		outPath := path.Join(outDir, filename)

		f, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			log.Println(err)
			continue
		}

		buf, err := generateTableAPI(tblDef)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = f.Write(buf)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func generateTableAPI(tbl tableDef) ([]byte, error) {
	buf := new(bytes.Buffer)

	generateHeader(tbl.Name, buf)
	generateAPIOnPrimIdx(tbl.Name, tbl.PrimaryIndex, buf)
	for idxName, sidx := range tbl.SecondaryIndexes {
		generateAPIOnScndIdx(tbl.Name, idxName, sidx, buf)
	}

	return format.Source(buf.Bytes())
}

func generateHeader(tblName string, buf io.Writer) {
	tblNameLCamel := strcase.ToLowerCamel(tblName)
	tblNameUCamel := strcase.ToCamel(tblName)

	fmt.Fprintf(buf, headerTemplate, tblNameLCamel, tblNameUCamel, tblName)
}

func generateAPIOnPrimIdx(tblName string, idx indexDef, buf io.Writer) {
	tblNameLCamel := strcase.ToLowerCamel(tblName)
	tblNameUCamel := strcase.ToCamel(tblName)

	if idx.isSimple() {
		pkNameLCamel := strcase.ToLowerCamel(idx.PK.Name)
		pkTypeName := attrType2GoType[idx.PK.Type]

		fmt.Fprintf(buf, withSimplePrimaryIndexTemplate, tblNameLCamel, tblNameUCamel, pkNameLCamel, pkTypeName, idx.PK.Name)
	} else {
		pkNameLCamel := strcase.ToLowerCamel(idx.PK.Name)
		pkTypeName := attrType2GoType[idx.PK.Type]
		skNameLCamel := strcase.ToLowerCamel(idx.SK.Name)
		skTypeName := attrType2GoType[idx.SK.Type]

		fmt.Fprintf(buf, withCompositePrimaryIndexTemplate, tblNameLCamel, tblNameUCamel, pkNameLCamel, pkTypeName, idx.PK.Name, skNameLCamel, skTypeName, idx.SK.Name)
	}
}

func generateAPIOnScndIdx(tblName string, idxName string, idx indexDef, buf io.Writer) {
	tblNameLCamel := strcase.ToLowerCamel(tblName)
	tblNameUCamel := strcase.ToCamel(tblName)
	idxNameUCamel := strcase.ToCamel(idxName)

	if idx.isSimple() {
		pkNameLCamel := strcase.ToLowerCamel(idx.PK.Name)
		pkTypeName := attrType2GoType[idx.PK.Type]

		fmt.Fprintf(buf, withSimpleSecondaryIndexTemplate, tblNameLCamel, tblNameUCamel, idxNameUCamel, idxName, pkNameLCamel, pkTypeName, idx.PK.Name)
	} else {
		pkNameLCamel := strcase.ToLowerCamel(idx.PK.Name)
		pkTypeName := attrType2GoType[idx.PK.Type]
		skNameLCamel := strcase.ToLowerCamel(idx.SK.Name)
		skTypeName := attrType2GoType[idx.SK.Type]

		fmt.Fprintf(buf, withCompositeSecondaryIndexTemplate, tblNameLCamel, tblNameUCamel, idxNameUCamel, idxName, pkNameLCamel, pkTypeName, idx.PK.Name, skNameLCamel, skTypeName, idx.SK.Name)
	}
}

// [1]: lower camel of tableName
// [2]: upper camel of tableName
// [3]: raw tableName
const headerTemplate = `
package ddbtbl

import "github.com/guregu/dynamo"

type %[1]sAPI struct {
	table dynamo.Table
}

// %[2]s is the entry point of manipulation of %[3]q.
func %[2]s(d *dynamo.DB) *%[1]sAPI {
	return &%[1]sAPI{table: d.Table(%[3]q)}
}

// index-free APIs
// Scan on %[3]q.
func (a *%[1]sAPI) Scan() *dynamo.Scan {
	return a.table.Scan()
}

// Put item to %[3]q.
func (a *%[1]sAPI) Put(item interface{}) *dynamo.Put {
	return a.table.Put(item)
}

// BatchPut puts items to %[3]q.
func (a *%[1]sAPI) BatchPut(items ...interface{}) *dynamo.BatchWrite {
	return a.table.Batch().Write().Put(items)
}
`

// [1]: lower camel of tableName
// [2]: upper camel of tableName
// [3]: lower camel of pkName
// [4]: type name of pkName
// [5]: raw pkName
const withSimplePrimaryIndexTemplate = `
// primary index API
func(a *%[1]sAPI) WithPrimaryIndex(%[3]s %[4]s) *withPrimIdx {
	return &withPrimIdx{
		table: a.table,
		pkName: %[5]q,
		pkVal: %[3]s,
	}
}

type %[2]sPrimIndex struct {
	%[3]s	%[4]s
}

func (a *%[1]sAPI) BatchWithPrimaryIndex(keys []%[2]sPrimIndex) *batchWithPrimIdx {
	ks := make([]dynamo.Keyed, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, dynamo.Keys{k.%[3]s, nil})
	}

	return &batchWithPrimIdx {
		table: a.table,
		pkName: %[5]q,
		keys: ks,
	}
}
`

// [1]: lower camel of tableName
// [2]: upper camel of tableName
//
// [3]: lower camel of pkName
// [4]: type name of pkName
// [5]: raw pkName
//
// [6]: lower camel of skName
// [7]: type name of skName
// [8]: raw skName
const withCompositePrimaryIndexTemplate = `
// primary index API
func(a *%[1]sAPI) WithPrimaryIndex(%[3]s %[4]s, %[6]s %[7]s) *withPrimIdx {
	return &withPrimIdx{
		table: a.table,
		pkName: %[5]q,
		pkVal: %[3]s,
		skName: %[8]q,
		skVal: %[6]s,
	}
}

func (a *%[1]sAPI) QueryWithPrimaryIndex(%[3]s %[4]s) *queryWithPrimIdx {
	return &queryWithPrimIdx{
		table: a.table,
		pkName: %[5]q,
		pkVal: %[3]s,
		skName: %[8]q,
	}
}

type %[2]sPrimIndex struct {
	%[3]s %[4]s
	%[6]s %[7]s
}

func (a *%[1]sAPI) BatchWithPrimaryIndex(keys []%[2]sPrimIndex) *batchWithPrimIdx {
	ks := make([]dynamo.Keyed, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, dynamo.Keys{k.%[3]s, k.%[6]s})
	}

	return &batchWithPrimIdx {
		table: a.table,
		pkName: %[5]q,
		skName: %[8]q,
		keys: ks,
	}
}
`

// [1]: lower camel of tableName
// [2]: upper camel of tableName
//
// [3]: upper camel of indexName
// [4]: raw indexName
//
// [5]: lower camel of pkName
// [6]: type name of pkName
// [7]: raw pkName
const withSimpleSecondaryIndexTemplate = `
// %[4]q (secondary index) API
func(a *%[1]sAPI) With%[3]s(%[5]s %[6]s) *withScndIdx {
	return &withScndIdx{
		table: a.table,
		idxName: %[4]q,
		pkName: %[7]q,
		pkVal: %[5]s,
	}
}

type %[2]s%[3]s struct {
	%[5]s	%[6]s
}

func (a *%[1]sAPI) BatchWith%[3]s(keys []%[2]s%[3]s) *batchWithScndIdx {
	ks := make([]dynamo.Keyed, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, dynamo.Keys{k.%[5]s, nil})
	}

	return &batchWithScndIdx {
		table: a.table,
		pkName: %[7]q,
		keys: ks,
	}
}
`

// [1]: lower camel of tableName
// [2]: upper camel of tableName
//
// [3]: upper camel of indexName
// [4]: raw indexName
//
// [5]: lower camel of pkName
// [6]: type name of pkName
// [7]: raw pkName
//
// [8]: lower camel of skName
// [9]: type name of skName
// [10]: raw skName
const withCompositeSecondaryIndexTemplate = `
// %[4]q (secondary index) API
func(a *%[1]sAPI) With%[3]s(%[5]s %[6]s, %[8]s %[9]s) *withScndIdx {
	return &withScndIdx{
		table: a.table,
		idxName: %[4]q,
		pkName: %[7]q,
		pkVal: %[5]s,
		skName: %[10]q,
		skVal: %[8]s,
	}
}

func (a *%[1]sAPI) QueryWith%[3]s(%[5]s %[6]s) *queryWithScndIdx {
	return &queryWithScndIdx{
		table: a.table,
		idxName: %[4]q,
		pkName: %[7]q,
		pkVal: %[5]s,
		skName: %[10]q,
	}
}

type %[2]s%[3]s struct {
	%[5]s %[6]s
	%[8]s %[9]s
}

func (a *%[1]sAPI) BatchWith%[3]s(keys []%[2]s%[3]s) *batchWithScndIdx {
	ks := make([]dynamo.Keyed, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, dynamo.Keys{k.%[5]s, k.%[8]s})
	}

	return &batchWithScndIdx {
		table: a.table,
		pkName: %[7]q,
		skName: %[10]q,
		keys: ks,
	}
}
`
