package ddbtbl

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
