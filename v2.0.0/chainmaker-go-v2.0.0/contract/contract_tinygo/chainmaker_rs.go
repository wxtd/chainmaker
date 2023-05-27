package main

// ResultSet iterator query result
type ResultSet interface {
	// NextRow get next row,
	// sql: column name is EasyCodec key, value is EasyCodec string val. as: val := ec.getString("columnName")
	// kv iterator: key/value is EasyCodec key for "key"/"value", value type is []byte. as: k, _ := ec.GetString("key") v, _ := ec.GetBytes("value")
	NextRow() (*EasyCodec, ResultCode)
	// HasNext return does the next line exist
	HasNext() bool
	// close
	Close() (bool, ResultCode)
}

type ResultSetKV interface {
	ResultSet
	// Next return key,field,value,code
	Next() (string, string, []byte, ResultCode)
}

type SqlSimContext interface {
	SimContextCommon
	// sql method
	// ExecuteQueryOne
	ExecuteQueryOne(sql string) (*EasyCodec, ResultCode)
	ExecuteQuery(sql string) (ResultSet, ResultCode)
	// #### ExecuteUpdateSql execute update/insert/delete sql
	// ##### It is best to update with primary key
	//
	// as:
	//
	// - update table set name = 'Tom' where uniqueKey='xxx'
	// - delete from table where uniqueKey='xxx'
	// - insert into table(id, xxx,xxx) values(xxx,xxx,xxx)
	//
	// ### not allow:
	// - random methods: NOW() RAND() and so on
	// return: 1 Number of rows affected;2 result code
	ExecuteUpdate(sql string) (int32, ResultCode)
	// ExecuteDDLSql execute DDL sql, for init_contract or upgrade method. allow table create/alter/drop/truncate
	//
	// ## You must have a primary key to create a table
	// ### allow:
	// - CREATE TABLE tableName
	// - ALTER TABLE tableName
	// - DROP TABLE tableName
	// - TRUNCATE TABLE tableName
	//
	// ### not allow:
	// - CREATE DATABASE dbName
	// - CREATE TABLE dbName.tableName
	// - ALTER TABLE dbName.tableName
	// - DROP DATABASE dbName
	// - DROP TABLE dbName.tableName
	// - TRUNCATE TABLE dbName.tableName
	// not allow:
	// - random methods: NOW() RAND() and so on
	//
	ExecuteDdl(sql string) (int32, ResultCode)
}

type SqlSimContextImpl struct {
	SimContextCommonImpl
}

func NewSqlSimContext() SqlSimContext {
	return &SqlSimContextImpl{}
}

// sql
func (s *SqlSimContextImpl) ExecuteQueryOne(sql string) (*EasyCodec, ResultCode) {
	ec := NewEasyCodec()
	ec.AddString("sql", sql)
	bytes, code := GetBytesFromChain(ec, ContractMethodExecuteQueryOneLen, ContractMethodExecuteQueryOne)
	if code == SUCCESS {
		return NewEasyCodecWithBytes(bytes), code
	}
	return NewEasyCodec(), ERROR
}

func (s *SqlSimContextImpl) ExecuteQuery(sql string) (ResultSet, ResultCode) {
	ec := NewEasyCodec()
	ec.AddString("sql", sql)
	index, code := GetInt32FromChain(ec, ContractMethodExecuteQuery)
	return NewResultSet(s, index), code
}

func (s *SqlSimContextImpl) ExecuteUpdate(sql string) (int32, ResultCode) {
	ec := NewEasyCodec()
	ec.AddString("sql", sql)
	return GetInt32FromChain(ec, ContractMethodExecuteUpdate)
}

func (s *SqlSimContextImpl) ExecuteDdl(sql string) (int32, ResultCode) {
	ec := NewEasyCodec()
	ec.AddString("sql", sql)
	return GetInt32FromChain(ec, ContractMethodExecuteDdl)
}

func (s *SqlSimContextImpl) IteratorNextRow(rsIndex int32) ([]byte, ResultCode) {
	ec := NewEasyCodec()
	ec.AddInt32("rs_index", rsIndex)
	return GetBytesFromChain(ec, ContractMethodRSNextLen, ContractMethodRSNext)
}

func (s *SqlSimContextImpl) IteratorHasNext(rsIndex int32) (int32, ResultCode) {
	ec := NewEasyCodec()
	ec.AddInt32("rs_index", rsIndex)
	return GetInt32FromChain(ec, ContractMethodRSHasNext)
}
func (s *SqlSimContextImpl) IteratorClose(rsIndex int32) (int32, ResultCode) {
	ec := NewEasyCodec()
	ec.AddInt32("rs_index", rsIndex)
	return GetInt32FromChain(ec, ContractMethodRSClose)
}

type ResultSetImpl struct {
	sqlCtx *SqlSimContextImpl
	index  int32 // 链的rs句柄的index
}

func NewResultSet(sqlCtx *SqlSimContextImpl, index int32) ResultSet {
	return &ResultSetImpl{sqlCtx, index}
}

func (r *ResultSetImpl) NextRow() (*EasyCodec, ResultCode) {
	bytes, code := r.sqlCtx.IteratorNextRow(r.index)
	if code != SUCCESS {
		return NewEasyCodec(), ERROR
	}
	return NewEasyCodecWithBytes(bytes), SUCCESS
}

func (r *ResultSetImpl) HasNext() bool {
	data, _ := r.sqlCtx.IteratorHasNext(r.index)
	return data != 0
}

func (r *ResultSetImpl) Close() (bool, ResultCode) {
	data, code := r.sqlCtx.IteratorClose(r.index)
	return data != 0, code
}
