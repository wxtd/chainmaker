/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sqldbprovider

import (
	"testing"
	"time"

	"chainmaker.org/chainmaker-go/localconf"
	"chainmaker.org/chainmaker/protocol/v2/test"
	"github.com/stretchr/testify/assert"
)

var log = &test.GoLogger{}

//func TestProvider_GetDB(t *testing.T) {
//	conf := &localconf.CMConfig{}
//	conf.StorageConfig.MysqlConfig.Dsn = "root:123456@tcp(127.0.0.1:3306)/"
//	conf.StorageConfig.MysqlConfig.MaxIdleConns = 10
//	conf.StorageConfig.MysqlConfig.MaxOpenConns = 10
//	conf.StorageConfig.MysqlConfig.ConnMaxLifeTime = 60
//	var db SqlDBHandle
//	db.GetDB("test_chain_1", conf)
//}
var conf = &localconf.SqlDbConfig{
	Dsn:        "file::memory:?cache=shared",
	SqlDbType:  "sqlite",
	SqlLogMode: "info",
}

func TestProvider_ExecSql(t *testing.T) {

	p := NewSqlDBHandle("chain1", conf, log)
	p.ExecSql("create table t1(id int primary key,name varchar(5))", "")
	p.ExecSql("insert into t1 values(1,'a')", "")

	count, err := p.ExecSql("update t1 set name='aa' where id=1", "")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	count, err = p.ExecSql("update t1 set name1='aa' where id=1", "")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), count)
}
func TestProvider_QuerySql(t *testing.T) {

	p := NewSqlDBHandle("chain1", conf, log)
	p.ExecSql("create table t1(id int primary key,name varchar(5))", "")
	p.ExecSql("insert into t1 values(1,'a')", "")
	p.ExecSql("insert into t1 values(2,'b')", "")
	row, err := p.QuerySingle("select count(*) from t1", "")
	assert.Nil(t, err)
	var id int
	err = row.ScanColumns(&id)
	assert.Nil(t, err)
	assert.Equal(t, 2, id)
	row, err = p.QuerySingle("select name from t1 where id=?", 3)
	assert.Nil(t, err)
	assert.True(t, row.IsEmpty())
	data, err := row.Data()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(data))
}
func TestProvider_QueryTableSql(t *testing.T) {

	p := NewSqlDBHandle("chain1", conf, log)
	p.ExecSql("create table t1(id int primary key,name varchar(5))", "")
	p.ExecSql("insert into t1 values(1,'a')", "")
	p.ExecSql("insert into t1 values(2,'b')", "")
	rows, err := p.QueryMulti("select * from t1", "")
	assert.Nil(t, err)
	defer rows.Close()
	var id int
	var name string
	for rows.Next() {
		rows.ScanColumns(&id, &name)
		t.Log(id, name)
	}
}
func initProvider() *SqlDBHandle {
	p := NewSqlDBHandle("chain1", conf, log)
	return p
}
func initData(p *SqlDBHandle) {
	p.ExecSql("create table t1(id int primary key,name varchar(5))", "")
	p.ExecSql("insert into t1 values(1,'a')", "")
	p.ExecSql("insert into t1 values(2,'b')", "")
}

func TestProvider_DbTransaction(t *testing.T) {
	p := initProvider()
	initData(p)
	txName := "Block1"
	tx, _ := p.BeginDbTransaction(txName)
	tx.BeginDbSavePoint("tx0")
	var count int64
	var err error
	count, _ = tx.ExecSql("insert into t1 values(3,'c')")
	assert.Equal(t, int64(1), count)
	count, _ = tx.ExecSql("insert into t1 values(4,'d')")
	assert.Equal(t, int64(1), count)
	tx.BeginDbSavePoint("tx1")
	count, _ = tx.ExecSql("insert into t1 values(5,'e')")
	assert.Equal(t, int64(1), count)
	row, _ := tx.QuerySingle("select count(*) from t1")
	row.ScanColumns(&count)
	assert.Equal(t, int64(5), count)
	count, err = tx.ExecSql("insert into t1 values(2,'b')") //duplicate PK error
	assert.NotNil(t, err)
	tx.RollbackDbSavePoint("tx1")
	row, err = tx.QuerySingle("select count(*) from t1")
	row.ScanColumns(&count)
	assert.Nil(t, err)
	assert.Equal(t, int64(4), count)
	p.RollbackDbTransaction(txName)
	row, err = p.QuerySingle("select count(1) from t1", "")
	assert.Nil(t, err)
	row.ScanColumns(&count)
	assert.Equal(t, int64(2), count)
}

func TestProvider_RollbackEmptyTx(t *testing.T) {
	p := initProvider()
	initData(p)
	txName := "Block1"
	tx, _ := p.BeginDbTransaction(txName)
	tx.BeginDbSavePoint("tx0")
	var count int64
	var err error
	count, err = tx.ExecSql("insert into t1 values(3,'c','error')")
	assert.NotNil(t, err)
	tx.RollbackDbSavePoint("tx0")
	row, err := tx.QuerySingle("select count(*) from t1")
	assert.Nil(t, err)
	row.ScanColumns(&count)
	assert.Equal(t, int64(2), count)
	p.RollbackDbTransaction(txName)
	row, err = p.QuerySingle("select count(1) from t1")
	assert.Nil(t, err)
	row.ScanColumns(&count)
	assert.Equal(t, int64(2), count)
}

func TestProvider_RollbackSavepointByInvalidSql(t *testing.T) {
	p := initProvider()
	initData(p)
	txName := "Block1"
	tx, _ := p.BeginDbTransaction(txName)
	tx.BeginDbSavePoint("tx0")
	var count int64
	var err error
	count, err = tx.ExecSql("insert into t1 values(3,'c')")
	assert.Nil(t, err)
	row, err := tx.QuerySingle("select count(*) from t1")
	row.ScanColumns(&count)
	assert.Equal(t, int64(3), count)
	count, err = tx.ExecSql("insert into t1 values(4,'cc")
	assert.NotNil(t, err)
	tx.RollbackDbSavePoint("tx0")
	row, err = tx.QuerySingle("select count(*) from t1")
	row.ScanColumns(&count)
	assert.Equal(t, int64(2), count)
	p.RollbackDbTransaction(txName)
	row, err = p.QuerySingle("select count(1) from t1")
	row.ScanColumns(&count)
	assert.Equal(t, int64(2), count)
}
func TestSqlDBHandle_QuerySql(t *testing.T) {
	p := initProvider()
	p.ExecSql("create table t1(id int primary key,name varchar(50),birthdate datetime,photo blob)", "")
	var bin = []byte{1, 2, 3, 4, 0xff}
	p.ExecSql("insert into t1 values(?,?,?,?)", 1, "Devin", time.Now(), bin)
	p.ExecSql("insert into t1 values(?,?,?,?)", 2, "Edward", time.Now(), bin)
	p.ExecSql("insert into t1 values(?,?,?,?)", 3, "Devin", time.Now(), bin)
	result, err := p.QuerySingle("select * from t1 where name=?", "Devin")
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(result.Data())
}
func TestReplaceDsn(t *testing.T) {
	tb := make(map[string]string)
	dbName := "blockdb_chain1"
	tb["root:123@456@tcp(127.0.0.1)/mysql?charset=utf8mb4&parseTime=True&loc=Local"] = "root:123@456@tcp(127.0.0.1)/blockdb_chain1?charset=utf8mb4&parseTime=True&loc=Local"
	tb["root:123@456@tcp(127.0.0.1)/{0}"] = "root:123@456@tcp(127.0.0.1)/blockdb_chain1"
	tb["root:123@456@tcp(127.0.0.1)/"] = "root:123@456@tcp(127.0.0.1)/blockdb_chain1"
	tb["root:123@mysql@tcp(127.0.0.1)/mysql"] = "root:123@mysql@tcp(127.0.0.1)/blockdb_chain1"
	tb["root:123@mysql@tcp(127.0.0.1)mysql"] = "root:123@mysql@tcp(127.0.0.1)mysql"
	for dsn, result := range tb {
		replaced := replaceMySqlDsn(dsn, dbName)
		assert.Equal(t, result, replaced)
	}
}
