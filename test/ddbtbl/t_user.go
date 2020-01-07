package ddbtbl

import "github.com/guregu/dynamo"

type tUserAPI struct {
	table dynamo.Table
}

// TUser is the entry point of manipulation of "TUser".
func TUser(d *dynamo.DB) *tUserAPI {
	return &tUserAPI{table: d.Table("TUser")}
}

// index-free APIs
// Scan on "TUser".
func (a *tUserAPI) Scan() *dynamo.Scan {
	return a.table.Scan()
}

// Put item to "TUser".
func (a *tUserAPI) Put(item interface{}) *dynamo.Put {
	return a.table.Put(item)
}

// BatchPut puts items to "TUser".
func (a *tUserAPI) BatchPut(items ...interface{}) *dynamo.BatchWrite {
	return a.table.Batch().Write().Put(items...)
}

// primary index API
func (a *tUserAPI) WithPrimaryIndex(uid int) *withPrimIdx {
	return &withPrimIdx{
		table:  a.table,
		pkName: "uid",
		pkVal:  uid,
	}
}

func (a *tUserAPI) BatchWithPrimaryIndex(uidList []int) *batchWithPrimIdx {
	dedup := make(map[int]struct{})
	for _, k := range uidList {
		dedup[k] = struct{}{}
	}

	ks := make([]dynamo.Keyed, 0, len(dedup))
	for k := range dedup {
		ks = append(ks, dynamo.Keys{k, nil})
	}

	return &batchWithPrimIdx{
		table:  a.table,
		pkName: "uid",
		keys:   ks,
	}
}

// "name-index" (secondary index) API
func (a *tUserAPI) WithNameIndex(name string) *withScndIdx {
	return &withScndIdx{
		table:   a.table,
		idxName: "name-index",
		pkName:  "name",
		pkVal:   name,
	}
}

func (a *tUserAPI) BatchWithNameIndex(nameList []string) *batchWithScndIdx {
	dedup := make(map[string]struct{})
	for _, k := range nameList {
		dedup[k] = struct{}{}
	}

	ks := make([]dynamo.Keyed, 0, len(dedup))
	for k := range dedup {
		ks = append(ks, dynamo.Keys{k, nil})
	}

	return &batchWithScndIdx{
		table:  a.table,
		pkName: "name",
		keys:   ks,
	}
}
