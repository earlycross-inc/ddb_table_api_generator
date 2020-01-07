package test

import (
	"testing"

	"bitbucket.org/earlycross/ddb_table_api_generator/test/ddbtbl"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

func initDDBCli() *dynamo.DB {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-1"),
		Endpoint:    aws.String("http://localhost:8100"),
		Credentials: credentials.NewStaticCredentials("dummy", "dummy", ""),
	}))

	return dynamo.New(sess)
}

type tUser struct {
	UID  int    `dynamo:"uid,hash"`
	Name string `dynamo:"name" index:"name-index,hash"`
}

func TestSimpleIndexAPI(t *testing.T) {
	ddbCli := initDDBCli()
	err := ddbCli.CreateTable("TUser", tUser{}).Run()
	if err != nil {
		t.Fatal("failed to create table TUser for test:", err)
	}

	tUserAPI := ddbtbl.TUser(ddbCli)

	u1 := tUser{
		UID:  1,
		Name: "hoge",
	}
	err = tUserAPI.Put(u1).Run()
	if err != nil {
		t.Fatal("failed to put: ", err)
	}

	var getU1 tUser
	err = tUserAPI.WithPrimaryIndex(u1.UID).Get().One(&getU1)
	if err != nil {
		t.Fatal("failed to get: ", err)
	}
	if getU1.Name != u1.Name {
		t.Fatalf("unexpected user name: %s", getU1.Name)
	}

	newU1Name := "fuga"
	err = tUserAPI.WithPrimaryIndex(u1.UID).Update().Set("name", newU1Name).Run()
	if err != nil {
		t.Fatal("failed to update: ", err)
	}
	err = tUserAPI.WithPrimaryIndex(u1.UID).Get().One(&getU1)
	if err != nil {
		t.Fatal("failed to get after update: ", err)
	}
	if getU1.Name != newU1Name {
		t.Fatalf("unexpected updated user name: %s", getU1.Name)
	}

	err = tUserAPI.WithPrimaryIndex(u1.UID).Delete().Run()
	if err != nil {
		t.Fatal("failed to delete: ", err)
	}
	err = tUserAPI.WithPrimaryIndex(u1.UID).Get().One(&getU1)
	if err == nil {
		t.Fatal("unexpectedly succeed to get deleted user")
	} else if err != dynamo.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}

	users := []interface{}{
		tUser{UID: 1, Name: "test1"},
		tUser{UID: 2, Name: "test2"},
		tUser{UID: 3, Name: "test3"},
	}
	_, err = tUserAPI.BatchPut(users...).Run()
	if err != nil {
		t.Fatal("failed to batchPut: ", err)
	}

	getUsers := make([]tUser, 0)
	err = tUserAPI.Scan().All(&getUsers)
	if err != nil {
		t.Fatal("failed to scan: ", err)
	}
	if len(getUsers) != len(users) {
		t.Fatalf("num of scanned users mismatch. got=%d", len(getUsers))
	}

	var getTest1 tUser
	err = tUserAPI.WithNameIndex("test1").Get().One(&getTest1)
	if err != nil {
		t.Fatal("failed to get with name-index: ", err)
	}
	if getTest1.UID != 1 {
		t.Fatalf("uid of test1 mismatch. got=%d", getTest1.UID)
	}

	targetUIDs := []int{1, 3, 3}
	getUsers = make([]tUser, 0)
	err = tUserAPI.BatchWithPrimaryIndex(targetUIDs).Get().All(&getUsers)
	if err != nil {
		t.Fatal("failed to batchGet: ", err)
	}
	if len(getUsers) != 2 {
		t.Fatalf("num of fetched users mismatch. got=%d, want=2", len(getUsers))
	}

	delUIDs := []int{1, 2}
	_, err = tUserAPI.BatchWithPrimaryIndex(delUIDs).Delete().
		Put(tUser{UID: 4, Name: "test4"}, tUser{UID: 5, Name: "test5"}).
		Run()
	if err != nil {
		t.Fatal("failed to batchWrite: ", err)
	}

	getUsers = make([]tUser, 0)
	err = tUserAPI.Scan().All(&getUsers)
	if err != nil {
		t.Fatal("failed to scan(after batchWrite): ", err)
	}
	if len(getUsers) != 3 {
		t.Fatalf("num of fetched users(after batchWrite) mismatch. got=%d, want=3", len(getUsers))
	}
}

type tUserStageRanking struct {
	UID   int    `dynamo:"uid,hash"`
	StgID int64  `dynamo:"stg_id,range"`
	Pkey  string `dynamo:"pkey" index:"pkey-index,hash"`
	Score int    `dynamo:"score" index:"pkey-index,range"`
}

func TestCompositeIndexAPI(t *testing.T) {
	ddbCli := initDDBCli()
	err := ddbCli.CreateTable("TUserStageRanking", tUserStageRanking{}).Run()
	if err != nil {
		t.Fatal("failed to create table TUserStageRanking for test: ", err)
	}

	tRankingAPI := ddbtbl.TUserStageRanking(ddbCli)

	record1 := tUserStageRanking{
		UID:   1,
		StgID: 2,
		Pkey:  "p-1",
		Score: 100,
	}
	_ = tRankingAPI.Put(record1).Run()

	var getRecord1 tUserStageRanking
	err = tRankingAPI.WithPrimaryIndex(record1.UID, record1.StgID).Get().One(&getRecord1)
	if err != nil {
		t.Fatal("failed to get: ", err)
	}
	if getRecord1.Score != record1.Score {
		t.Fatalf("unexpected score: %d", getRecord1.Score)
	}

	err = tRankingAPI.WithPkeyIndex(record1.Pkey, record1.Score).Get().One(&getRecord1)
	if err != nil {
		t.Fatal("failed to get by pkey index: ", err)
	}
	if getRecord1.UID != record1.UID {
		t.Fatalf("unexpected uid: %d", getRecord1.UID)
	}

	newScore := 200
	err = tRankingAPI.WithPrimaryIndex(record1.UID, record1.StgID).Update().Set("score", newScore).Run()
	if err != nil {
		t.Fatal("failed to update: ", err)
	}
	err = tRankingAPI.WithPrimaryIndex(record1.UID, record1.StgID).Get().One(&getRecord1)
	if err != nil {
		t.Fatal("failed to get after update: ", err)
	}
	if getRecord1.Score != newScore {
		t.Fatalf("unexpected updated score: %d", getRecord1.Score)
	}

	err = tRankingAPI.WithPrimaryIndex(record1.UID, record1.StgID).Delete().Run()
	if err != nil {
		t.Fatal("failed to delete: ", err)
	}
	err = tRankingAPI.WithPrimaryIndex(record1.UID, record1.StgID).Get().One(&getRecord1)
	if err == nil {
		t.Fatal("unexpectedly succeed to get deleted record")
	} else if err != dynamo.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}

	records := []interface{}{
		tUserStageRanking{UID: 11, StgID: 1, Pkey: "p-1", Score: 100},
		tUserStageRanking{UID: 21, StgID: 1, Pkey: "p-1", Score: 120},
		tUserStageRanking{UID: 12, StgID: 1, Pkey: "p-2", Score: 200},
		tUserStageRanking{UID: 22, StgID: 1, Pkey: "p-2", Score: 310},
		tUserStageRanking{UID: 22, StgID: 2, Pkey: "p-2", Score: 150},
		tUserStageRanking{UID: 22, StgID: 3, Pkey: "p-2", Score: 190},
	}
	_, _ = tRankingAPI.BatchPut(records...).Run()

	getRecords := make([]tUserStageRanking, 0)
	err = tRankingAPI.QueryWithPrimaryIndex(22).WhereSK(dynamo.LessOrEqual, 2).All(&getRecords)
	if err != nil {
		t.Fatal("failed to query with uid: ", err)
	}
	if len(getRecords) != 2 {
		t.Fatalf("num of fetched records by query(with uid) mismatch. got=%d, want=2", len(getRecords))
	}

	getRecords = make([]tUserStageRanking, 0)
	err = tRankingAPI.QueryWithPkeyIndex("p-2").All().All(&getRecords)
	if err != nil {
		t.Fatal("failed to query with pkey: ", err)
	}
	if len(getRecords) != 4 {
		t.Fatalf("num of fetched records by query(with pkey) mismatch. got=%d, want=4", len(getRecords))
	}

	primIdxKeys := []ddbtbl.TUserStageRankingPrimIndex{
		{Uid: 11, StgId: 1},
		{Uid: 12, StgId: 1},
		{Uid: 22, StgId: 1},
		{Uid: 12, StgId: 1},
	}
	getRecords = make([]tUserStageRanking, 0)
	err = tRankingAPI.BatchWithPrimaryIndex(primIdxKeys).Get().All(&getRecords)
	if err != nil {
		t.Fatal("failed to batchGet with primary key: ", err)
	}
	if len(getRecords) != 3 {
		t.Fatalf("num of fetched records by bacthGet(with prim key) mismatch. got=%d, want=3", len(getRecords))
	}
}
