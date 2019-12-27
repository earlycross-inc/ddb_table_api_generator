package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
)

func generateAll(defFilename string, outDir string) {
	err := os.MkdirAll(outDir, 0666)
	if err != nil {
		log.Println(err)
		return
	}

	// write a source file defines logics used by DDB table APIs
	ddbapiSrcPath := path.Join(outDir, "ddbapi.go")
	err = ioutil.WriteFile(ddbapiSrcPath, []byte(ddbapiSource), 0666)
	if err != nil {
		log.Println(err)
		return
	}

	// read table definition file
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

	// generate table API for each table definition
	for _, tblDef := range defs {
		if !tblDef.isValid() {
			log.Printf("invalid table definition. table name: %s", tblDef.Name)
			continue
		}

		buf, err := generateTableAPI(tblDef)
		if err != nil {
			log.Println(err)
			continue
		}

		filename := fmt.Sprintf("%s.go", strcase.ToSnake(tblDef.Name))
		outPath := path.Join(outDir, filename)
		err = ioutil.WriteFile(outPath, buf, 0666)
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
		pkNameUCamel := strcase.ToCamel(idx.PK.Name)
		pkTypeName := attrType2GoType[idx.PK.Type]

		skNameLCamel := strcase.ToLowerCamel(idx.SK.Name)
		skNameUCamel := strcase.ToCamel(idx.SK.Name)
		skTypeName := attrType2GoType[idx.SK.Type]

		fmt.Fprintf(buf, withCompositePrimaryIndexTemplate,
			tblNameLCamel, tblNameUCamel,
			pkNameLCamel, pkNameUCamel, pkTypeName, idx.PK.Name,
			skNameLCamel, skNameUCamel, skTypeName, idx.SK.Name)
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
		pkNameUCamel := strcase.ToCamel(idx.PK.Name)
		pkTypeName := attrType2GoType[idx.PK.Type]

		skNameLCamel := strcase.ToLowerCamel(idx.SK.Name)
		skNameUCamel := strcase.ToCamel(idx.SK.Name)
		skTypeName := attrType2GoType[idx.SK.Type]

		fmt.Fprintf(buf, withCompositeSecondaryIndexTemplate,
			tblNameLCamel, tblNameUCamel, idxNameUCamel, idxName,
			pkNameLCamel, pkNameUCamel, pkTypeName, idx.PK.Name,
			skNameLCamel, skNameUCamel, skTypeName, idx.SK.Name)
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

func (a *%[1]sAPI) BatchWithPrimaryIndex(%[3]sList []%[4]s) *batchWithPrimIdx {
	ks := make([]dynamo.Keyed, 0, len(%[3]sList))
	for _, %[3]s := range %[3]sList {
		ks = append(ks, dynamo.Keys{%[3]s, nil})
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
// [4]: upper camel of pkName
// [5]: type name of pkName
// [6]: raw pkName
//
// [7]: lower camel of skName
// [8]: upper camel of skName
// [9]: type name of skName
// [10]: raw skName
const withCompositePrimaryIndexTemplate = `
// primary index API
func(a *%[1]sAPI) WithPrimaryIndex(%[3]s %[5]s, %[7]s %[9]s) *withPrimIdx {
	return &withPrimIdx{
		table: a.table,
		pkName: %[6]q,
		pkVal: %[3]s,
		skName: %[10]q,
		skVal: %[7]s,
	}
}

func (a *%[1]sAPI) QueryWithPrimaryIndex(%[3]s %[5]s) *queryWithPrimIdx {
	return &queryWithPrimIdx{
		table: a.table,
		pkName: %[6]q,
		pkVal: %[3]s,
		skName: %[10]q,
	}
}

type %[2]sPrimIndex struct {
	%[4]s %[5]s
	%[8]s %[9]s
}

func (a *%[1]sAPI) BatchWithPrimaryIndex(keys []%[2]sPrimIndex) *batchWithPrimIdx {
	ks := make([]dynamo.Keyed, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, dynamo.Keys{k.%[4]s, k.%[8]s})
	}

	return &batchWithPrimIdx {
		table: a.table,
		pkName: %[6]q,
		skName: %[10]q,
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

func (a *%[1]sAPI) BatchWith%[3]s(%[5]sList []%[6]s) *batchWithScndIdx {
	ks := make([]dynamo.Keyed, 0, len(%[5]sList))
	for _, %[5]s := range %[5]sList {
		ks = append(ks, dynamo.Keys{%[5]s, nil})
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
// [6]: upper camel of pkName
// [7]: type name of pkName
// [8]: raw pkName
//
// [9]: lower camel of skName
// [10]: upper camel of skName
// [11]: type name of skName
// [12]: raw skName
const withCompositeSecondaryIndexTemplate = `
// %[4]q (secondary index) API
func(a *%[1]sAPI) With%[3]s(%[5]s %[7]s, %[9]s %[11]s) *withScndIdx {
	return &withScndIdx{
		table: a.table,
		idxName: %[4]q,
		pkName: %[8]q,
		pkVal: %[5]s,
		skName: %[12]q,
		skVal: %[9]s,
	}
}

func (a *%[1]sAPI) QueryWith%[3]s(%[5]s %[7]s) *queryWithScndIdx {
	return &queryWithScndIdx{
		table: a.table,
		idxName: %[4]q,
		pkName: %[8]q,
		pkVal: %[5]s,
		skName: %[12]q,
	}
}

type %[2]s%[3]s struct {
	%[6]s %[7]s
	%[10]s %[11]s
}

func (a *%[1]sAPI) BatchWith%[3]s(keys []%[2]s%[3]s) *batchWithScndIdx {
	ks := make([]dynamo.Keyed, 0, len(keys))
	for _, k := range keys {
		ks = append(ks, dynamo.Keys{k.%[6]s, k.%[10]s})
	}

	return &batchWithScndIdx {
		table: a.table,
		pkName: %[8]q,
		skName: %[12]q,
		keys: ks,
	}
}
`

const ddbapiSource = `package ddbtbl

import (
	"github.com/guregu/dynamo"
)

type withPrimIdx struct {
	table dynamo.Table

	pkName string
	pkVal  interface{}
	skName string
	skVal  interface{}
}

func (i *withPrimIdx) Get() *dynamo.Query {
	q := i.table.Get(i.pkName, i.pkVal)
	if i.skName != "" {
		q = q.Range(i.skName, dynamo.Equal, i.skVal)
	}
	return q
}

func (i *withPrimIdx) Update() *dynamo.Update {
	u := i.table.Update(i.pkName, i.pkVal)
	if i.skName != "" {
		u = u.Range(i.skName, i.skVal)
	}
	return u
}

func (i *withPrimIdx) Delete() *dynamo.Delete {
	d := i.table.Delete(i.pkName, i.pkVal)
	if i.skName != "" {
		d = d.Range(i.skName, i.skVal)
	}
	return d
}

type withScndIdx struct {
	table   dynamo.Table
	idxName string

	pkName string
	pkVal  interface{}
	skName string
	skVal  interface{}
}

func (i *withScndIdx) Get() *dynamo.Query {
	q := i.table.Get(i.pkName, i.pkVal).Index(i.idxName)
	if i.skName != "" {
		q = q.Range(i.skName, dynamo.Equal, i.skVal)
	}
	return q
}

type queryWithPrimIdx struct {
	table dynamo.Table

	pkName string
	pkVal  interface{}

	skName string
}

func (i *queryWithPrimIdx) All() *dynamo.Query {
	return i.table.Get(i.pkName, i.pkVal)
}

func (i *queryWithPrimIdx) WhereSK(op dynamo.Operator, sk interface{}) *dynamo.Query {
	return i.table.Get(i.pkName, i.pkVal).Range(i.skName, op, sk)
}

type queryWithScndIdx struct {
	table   dynamo.Table
	idxName string

	pkName string
	pkVal  interface{}
	skName string
}

func (i *queryWithScndIdx) All() *dynamo.Query {
	return i.table.Get(i.pkName, i.pkVal).Index(i.idxName)
}

func (i *queryWithScndIdx) WhereSK(op dynamo.Operator, sk interface{}) *dynamo.Query {
	return i.table.Get(i.pkName, i.pkVal).Range(i.skName, op, sk).Index(i.idxName)
}

type batchWithPrimIdx struct {
	table dynamo.Table

	pkName string
	skName string
	keys   []dynamo.Keyed
}

func (i *batchWithPrimIdx) Get() *dynamo.BatchGet {
	var batch dynamo.Batch
	if i.skName == "" {
		batch = i.table.Batch(i.pkName)
	} else {
		batch = i.table.Batch(i.pkName, i.skName)
	}

	return batch.Get(i.keys...)
}

func (i *batchWithPrimIdx) Delete() *dynamo.BatchWrite {
	var batch dynamo.Batch
	if i.skName == "" {
		batch = i.table.Batch(i.pkName)
	} else {
		batch = i.table.Batch(i.pkName, i.skName)
	}

	return batch.Write().Delete(i.keys...)
}

type batchWithScndIdx struct {
	table dynamo.Table

	pkName string
	skName string
	keys   []dynamo.Keyed
}

func (i *batchWithScndIdx) Get() *dynamo.BatchGet {
	var batch dynamo.Batch
	if i.skName == "" {
		batch = i.table.Batch(i.pkName)
	} else {
		batch = i.table.Batch(i.pkName, i.skName)
	}

	return batch.Get(i.keys...)
}
`
