package ddbtbl

import "github.com/guregu/dynamo"

type tUserStageRankingAPI struct {
	table dynamo.Table
}

// TUserStageRanking is the entry point of manipulation of "TUserStageRanking".
func TUserStageRanking(d *dynamo.DB) *tUserStageRankingAPI {
	return &tUserStageRankingAPI{table: d.Table("TUserStageRanking")}
}

// index-free APIs
// Scan on "TUserStageRanking".
func (a *tUserStageRankingAPI) Scan() *dynamo.Scan {
	return a.table.Scan()
}

// Put item to "TUserStageRanking".
func (a *tUserStageRankingAPI) Put(item interface{}) *dynamo.Put {
	return a.table.Put(item)
}

// BatchPut puts items to "TUserStageRanking".
func (a *tUserStageRankingAPI) BatchPut(items ...interface{}) *dynamo.BatchWrite {
	return a.table.Batch().Write().Put(items)
}

// primary index API
func (a *tUserStageRankingAPI) WithPrimaryIndex(uid int, stgId int64) *withPrimIdx {
	return &withPrimIdx{
		table:  a.table,
		pkName: "uid",
		pkVal:  uid,
		skName: "stg_id",
		skVal:  stgId,
	}
}

func (a *tUserStageRankingAPI) QueryWithPrimaryIndex(uid int) *queryWithPrimIdx {
	return &queryWithPrimIdx{
		table:  a.table,
		pkName: "uid",
		pkVal:  uid,
		skName: "stg_id",
	}
}

type TUserStageRankingPrimIndex struct {
	Uid   int
	StgId int64
}

func (a *tUserStageRankingAPI) BatchWithPrimaryIndex(keys []TUserStageRankingPrimIndex) *batchWithPrimIdx {
	dedup := make(map[TUserStageRankingPrimIndex]struct{})
	for _, k := range keys {
		dedup[k] = struct{}{}
	}

	ks := make([]dynamo.Keyed, 0, len(dedup))
	for k := range dedup {
		ks = append(ks, dynamo.Keys{k.Uid, k.StgId})
	}

	return &batchWithPrimIdx{
		table:  a.table,
		pkName: "uid",
		skName: "stg_id",
		keys:   ks,
	}
}

// "pkey-index" (secondary index) API
func (a *tUserStageRankingAPI) WithPkeyIndex(pkey string, score int) *withScndIdx {
	return &withScndIdx{
		table:   a.table,
		idxName: "pkey-index",
		pkName:  "pkey",
		pkVal:   pkey,
		skName:  "score",
		skVal:   score,
	}
}

func (a *tUserStageRankingAPI) QueryWithPkeyIndex(pkey string) *queryWithScndIdx {
	return &queryWithScndIdx{
		table:   a.table,
		idxName: "pkey-index",
		pkName:  "pkey",
		pkVal:   pkey,
		skName:  "score",
	}
}

type TUserStageRankingPkeyIndex struct {
	Pkey  string
	Score int
}

func (a *tUserStageRankingAPI) BatchWithPkeyIndex(keys []TUserStageRankingPkeyIndex) *batchWithScndIdx {
	dedup := make(map[TUserStageRankingPkeyIndex]struct{})
	for _, k := range keys {
		dedup[k] = struct{}{}
	}

	ks := make([]dynamo.Keyed, 0, len(dedup))
	for k := range dedup {
		ks = append(ks, dynamo.Keys{k.Pkey, k.Score})
	}

	return &batchWithScndIdx{
		table:  a.table,
		pkName: "pkey",
		skName: "score",
		keys:   ks,
	}
}
