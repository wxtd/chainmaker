package main

// sdk for user

import (
	"chainmaker.org/contract-sdk-tinygo/sdk/convert"
	"unsafe"
)

type ResultCode int

const (
	// special parameters passed to contract
	ContractParamCreatorOrgId = "__creator_org_id__"
	ContractParamCreatorRole  = "__creator_role__"
	ContractParamCreatorPk    = "__creator_pk__"
	ContractParamSenderOrgId  = "__sender_org_id__"
	ContractParamSenderRole   = "__sender_role__"
	ContractParamSenderPk     = "__sender_pk__"
	ContractParamBlockHeight  = "__block_height__"
	ContractParamTxId         = "__tx_id__"
	ContractParamContextPtr   = "__context_ptr__"

	// method name used by smart contract sdk
	// common
	ContractMethodLogMessage      = "LogMessage"
	ContractMethodSuccessResult   = "SuccessResult"
	ContractMethodErrorResult     = "ErrorResult"
	ContractMethodCallContract    = "CallContract"
	ContractMethodCallContractLen = "CallContractLen"
	ContractMethodEmitEvent       = "EmitEvent"
	// paillier
	ContractMethodGetPaillierOperationResult    = "GetPaillierOperationResult"
	ContractMethodGetPaillierOperationResultLen = "GetPaillierOperationResultLen"
	// bulletproofs
	ContractMethodGetBulletproofsResult    = "GetBulletproofsResult"
	ContractMethodGetBulletproofsResultLen = "GetBulletproofsResultLen"

	// kv
	ContractMethodGetStateLen = "GetStateLen"
	ContractMethodGetState    = "GetState"
	ContractMethodPutState    = "PutState"
	ContractMethodDeleteState = "DeleteState"
	// kv iterator
	ContractMethodKvIterator        = "KvIterator"
	ContractMethodKvPreIterator     = "KvPreIterator"
	ContractMethodKvIteratorHasNext = "KvIteratorHasNext"
	ContractMethodKvIteratorNextLen = "KvIteratorNextLen"
	ContractMethodKvIteratorNext    = "KvIteratorNext"
	ContractMethodKvIteratorClose   = "KvIteratorClose"
	// sql
	ContractMethodExecuteQuery       = "ExecuteQuery"
	ContractMethodExecuteQueryOne    = "ExecuteQueryOne"
	ContractMethodExecuteQueryOneLen = "ExecuteQueryOneLen"
	ContractMethodRSNext             = "RSNext"
	ContractMethodRSNextLen          = "RSNextLen"
	ContractMethodRSHasNext          = "RSHasNext"
	ContractMethodRSClose            = "RSClose"
	ContractMethodExecuteUpdate      = "ExecuteUpdate"
	ContractMethodExecuteDdl         = "ExecuteDDL"

	SUCCESS ResultCode = 0
	ERROR   ResultCode = 1
)

// sysCall provides data interaction with the chain. sysCallReq common param, request var param
//export sys_call
func sysCall(requestHeader string, requestBody string) int32

//export log_message
func logMessage(msg string)

// SimContextCommon common context
type SimContextCommon interface {
	// Arg get arg from transaction parameters, as:  arg1, code := ctx.Arg("arg1")
	Arg(key string) ([]byte, ResultCode)
	// Arg get arg from transaction parameters, as:  arg1, code := ctx.ArgString("arg1")
	ArgString(key string) (string, ResultCode)
	// Args return args
	Args() []*EasyCodecItem
	// Log record log to chain server
	Log(msg string)
	// SuccessResult record the execution result of the transaction, multiple calls will override
	SuccessResult(msg string)
	// SuccessResultByte record the execution result of the transaction, multiple calls will override
	SuccessResultByte(msg []byte)
	// ErrorResult record the execution result of the transaction. multiple calls will append. Once there is an error, it cannot be called success method
	ErrorResult(msg string)
	// CallContract cross contract call
	CallContract(contractName string, method string, param map[string][]byte) ([]byte, ResultCode)
	// GetCreatorOrgId get tx creator org id
	GetCreatorOrgId() (string, ResultCode)
	// GetCreatorRole get tx creator role
	GetCreatorRole() (string, ResultCode)
	// GetCreatorPk get tx creator pk
	GetCreatorPk() (string, ResultCode)
	// GetSenderOrgId get tx sender org id
	GetSenderOrgId() (string, ResultCode)
	// GetSenderOrgId get tx sender role
	GetSenderRole() (string, ResultCode)
	// GetSenderOrgId get tx sender pk
	GetSenderPk() (string, ResultCode)
	// GetBlockHeight get tx block height
	GetBlockHeight() (string, ResultCode)
	// GetTxId get current tx id
	GetTxId() (string, ResultCode)
	// EmitEvent emit event, you can subscribe to the event using the SDK
	EmitEvent(topic string, data ...string) ResultCode
}

// SimContext kv context
type SimContext interface {
	SimContextCommon
	// GetState get [key+"#"+field] from chain and db
	GetState(key string, field string) (string, ResultCode)
	// GetStateByte get [key+"#"+field] from chain and db
	GetStateByte(key string, field string) ([]byte, ResultCode)
	// GetStateByte get [key] from chain and db
	GetStateFromKey(key string) ([]byte, ResultCode)
	// PutState put [key+"#"+field, value] to chain
	PutState(key string, field string, value string) ResultCode
	// PutStateByte put [key+"#"+field, value] to chain
	PutStateByte(key string, field string, value []byte) ResultCode
	// PutStateFromKey put [key, value] to chain
	PutStateFromKey(key string, value string) ResultCode
	// PutStateFromKeyByte put [key, value] to chain
	PutStateFromKeyByte(key string, value []byte) ResultCode
	// DeleteState delete [key+"#"+field] to chain
	DeleteState(key string, field string) ResultCode
	// DeleteStateFromKey delete [key] to chain
	DeleteStateFromKey(key string) ResultCode
	// NewIterator range of [startKey, limitKey), front closed back open
	NewIterator(startKey string, limitKey string) (ResultSetKV, ResultCode)
	// NewIteratorWithField range of [key+"#"+startField, key+"#"+limitField), front closed back open
	NewIteratorWithField(key string, startField string, limitField string) (ResultSetKV, ResultCode)
	// NewIteratorPrefixWithKeyField range of [key+"#"+field, key+"#"+field], front closed back closed
	NewIteratorPrefixWithKeyField(key string, field string) (ResultSetKV, ResultCode)
	// NewIteratorPrefixWithKey range of [key, key], front closed back closed
	NewIteratorPrefixWithKey(key string) (ResultSetKV, ResultCode)
}

type SimContextCommonImpl struct {
}

type SimContextImpl struct {
	SimContextCommonImpl
}

func NewSimContext() SimContext {
	return &SimContextImpl{}
}
func (s *SimContextImpl) GetState(key string, field string) (string, ResultCode) {
	return GetState(key, field)
}
func (s *SimContextImpl) GetStateByte(key string, field string) ([]byte, ResultCode) {
	return GetStateByte(key, field)
}
func (s *SimContextImpl) GetStateFromKey(key string) ([]byte, ResultCode) {
	return GetStateByte(key, "")
}
func (s *SimContextImpl) PutState(key string, field string, value string) ResultCode {
	return PutState(key, field, value)
}
func (s *SimContextImpl) PutStateByte(key string, field string, value []byte) ResultCode {
	return PutState(key, field, string(value))
}
func (s *SimContextImpl) PutStateFromKey(key string, value string) ResultCode {
	return PutState(key, "", value)
}
func (s *SimContextImpl) PutStateFromKeyByte(key string, value []byte) ResultCode {
	return PutStateByte(key, "", value)
}
func (s *SimContextImpl) DeleteState(key string, field string) ResultCode {
	return DeleteState(key, field)
}
func (s *SimContextImpl) DeleteStateFromKey(key string) ResultCode {
	return DeleteState(key, "")
}

// common
func (s *SimContextCommonImpl) Arg(key string) ([]byte, ResultCode) {
	return Arg(key)
}
func (s *SimContextCommonImpl) ArgString(key string) (string, ResultCode) {
	val, code := Arg(key)
	return string(val), code
}
func (s *SimContextCommonImpl) Args() []*EasyCodecItem {
	return Args()
}
func (s *SimContextCommonImpl) Log(msg string) {
	LogMessage(msg)
}
func (s *SimContextCommonImpl) CallContract(contractName string, method string, param map[string][]byte) ([]byte, ResultCode) {
	return CallContract(contractName, method, param)
}
func (s *SimContextCommonImpl) SuccessResult(msg string) {
	sysCall(getRequestHeader(ContractMethodSuccessResult), msg)
}
func (s *SimContextCommonImpl) SuccessResultByte(msg []byte) {
	sysCall(getRequestHeader(ContractMethodSuccessResult), string(msg))
}
func (s *SimContextCommonImpl) ErrorResult(msg string) {
	sysCall(getRequestHeader(ContractMethodErrorResult), string(msg))
}
func (s *SimContextCommonImpl) GetCreatorOrgId() (string, ResultCode) {
	return stringArg(ContractParamCreatorOrgId)
}
func (s *SimContextCommonImpl) GetCreatorRole() (string, ResultCode) {
	return stringArg(ContractParamCreatorRole)
}
func (s *SimContextCommonImpl) GetCreatorPk() (string, ResultCode) {
	return stringArg(ContractParamCreatorPk)
}
func (s *SimContextCommonImpl) GetSenderOrgId() (string, ResultCode) {
	return stringArg(ContractParamSenderOrgId)
}
func (s *SimContextCommonImpl) GetSenderRole() (string, ResultCode) {
	return stringArg(ContractParamSenderRole)
}
func (s *SimContextCommonImpl) GetSenderPk() (string, ResultCode) {
	return stringArg(ContractParamSenderPk)
}
func (s *SimContextCommonImpl) GetBlockHeight() (string, ResultCode) {
	return stringArg(ContractParamBlockHeight)
}
func (s *SimContextCommonImpl) GetTxId() (string, ResultCode) {
	return stringArg(ContractParamTxId)
}
func (s *SimContextCommonImpl) EmitEvent(topic string, data ...string) ResultCode {
	return EmitEvent(topic, data...)
}

var argsBytes []byte
var argsMap []*EasyCodecItem
var argsFlag bool

//export runtime_type
func runtimeType() int32 {
	var ContractRuntimeGoSdkType int32 = 4
	argsFlag = false
	return ContractRuntimeGoSdkType
}

//export deallocate
func deallocate(size int32) {
	argsBytes = make([]byte, size)
	argsMap = make([]*EasyCodecItem, 0)
	argsFlag = false
}

//export allocate
func allocate(size int32) uintptr {
	argsBytes = make([]byte, size)
	argsMap = make([]*EasyCodecItem, 0)
	argsFlag = false

	return uintptr(unsafe.Pointer(&argsBytes[0]))
}

func getRequestHeader(method string) string {
	ec := NewEasyCodec()
	ec.AddValue(EasyKeyType_SYSTEM, "ctx_ptr", EasyValueType_INT32, getCtxPtr())
	ec.AddValue(EasyKeyType_SYSTEM, "version", EasyValueType_STRING, "v1.2.0")
	ec.AddValue(EasyKeyType_SYSTEM, "method", EasyValueType_STRING, method)
	return string(ec.Marshal())
}

// LogMessage
func LogMessage(msg string) {
	logMessage(msg)
}

// GetState get state from chain
func GetState(key string, field string) (string, ResultCode) {
	result, code := GetStateByte(key, field)
	if code != SUCCESS {
		return "", code
	}
	return string(result), code
}

// GetState get state from chain
func GetStateByte(key string, field string) ([]byte, ResultCode) {
	ec := NewEasyCodec()
	ec.AddString("key", key)
	ec.AddString("field", field)
	return GetBytesFromChain(ec, ContractMethodGetStateLen, ContractMethodGetState)
}

func GetBytesFromChain(ec *EasyCodec, methodLen string, method string) ([]byte, ResultCode) {
	// # get len
	// ## prepare param
	var valueLen int32 = 0
	valuePtr := int32(uintptr(unsafe.Pointer(&valueLen)))
	ec.AddInt32("value_ptr", valuePtr)
	b := ec.Marshal()
	// ## send req get len
	code := sysCall(getRequestHeader(methodLen), string(b))
	// ## verify
	if code != int32(SUCCESS) {
		return nil, ERROR
	}
	if valueLen == 0 {
		return nil, SUCCESS
	}
	// # get data
	// ## prepare param
	valueByte := make([]byte, valueLen)
	ec.RemoveKey("value_ptr")
	valuePtr = int32(uintptr(unsafe.Pointer(&valueByte[0])))
	ec.AddInt32("value_ptr", valuePtr)
	b = ec.Marshal()
	// ## send req get value
	code2 := sysCall(getRequestHeader(method), string(b))
	if code2 != int32(SUCCESS) {
		return nil, ERROR
	}
	return valueByte, SUCCESS
}

// GetInt32FromChain get i32 from chain
func GetInt32FromChain(ec *EasyCodec, method string) (int32, ResultCode) {
	// # get len
	// ## prepare param
	var valueLen int32 = 0
	valuePtr := int32(uintptr(unsafe.Pointer(&valueLen)))
	ec.AddInt32("value_ptr", valuePtr)
	b := ec.Marshal()
	// ## send req get len
	code := sysCall(getRequestHeader(method), string(b))
	return valueLen, ResultCode(code)
}

// GetStateFromKey get state from chain
func GetStateFromKey(key string) ([]byte, ResultCode) {
	return GetStateByte(key, "")
}

//EmitEvent emit Event to chain
func EmitEvent(topic string, data ...string) ResultCode {
	// prepare param
	var items []*EasyCodecItem
	items = make([]*EasyCodecItem, 0)
	items = append(items, &EasyCodecItem{
		KeyType:   EasyKeyType_USER,
		Key:       "topic",
		ValueType: EasyValueType_STRING,
		Value:     topic,
	})
	for index, value := range data {
		items = append(items, &EasyCodecItem{
			KeyType:   EasyKeyType_USER,
			Key:       "data" + convert.Int32ToString(int32(index)),
			ValueType: EasyValueType_STRING,
			Value:     value,
		})
	}
	b := EasyMarshal(items)
	reqBody := string(b)
	// send req put value
	code := sysCall(getRequestHeader(ContractMethodEmitEvent), reqBody)
	if code != int32(SUCCESS) {
		return ERROR
	}
	return SUCCESS
}

// PutState put state to chain
func PutState(key string, field string, value string) ResultCode {
	// prepare param
	ec := NewEasyCodec()
	ec.AddString("key", key)
	ec.AddString("field", field)
	ec.AddBytes("value", []byte(value))
	b := ec.Marshal()
	// send req put value
	code := sysCall(getRequestHeader(ContractMethodPutState), string(b))
	if code != int32(SUCCESS) {
		return ERROR
	}
	return SUCCESS
}

// PutState put state to chain
func PutStateByte(key string, field string, value []byte) ResultCode {
	return PutState(key, field, string(value))
}

// PutStateFromKey put state to chain
func PutStateFromKey(key string, value string) ResultCode {
	return PutState(key, "", value)
}

// PutStateFromKey put state to chain
func PutStateFromKeyByte(key string, value []byte) ResultCode {
	return PutStateByte(key, "", value)
}

// DeleteState delete state to chain
func DeleteState(key string, field string) ResultCode {
	// prepare param
	ec := NewEasyCodec()
	ec.AddString("key", key)
	ec.AddString("field", field)
	b := ec.Marshal()
	// send req put value
	code := sysCall(getRequestHeader(ContractMethodDeleteState), string(b))
	if code != int32(SUCCESS) {
		return ERROR
	}
	return SUCCESS
}

// CallContract call other contract from chain
func CallContract(contractName string, method string, param map[string][]byte) ([]byte, ResultCode) {
	// # get len
	// ## prepare param
	var valueLen int32 = 0
	valuePtr := int32(uintptr(unsafe.Pointer(&valueLen)))

	ec := NewEasyCodec()
	ecMap := NewEasyCodecWithMap(param)
	paramBytes := ecMap.Marshal()
	ec.AddBytes("param", paramBytes)
	ec.AddInt32("value_ptr", valuePtr)
	ec.AddString("contract_name", contractName)
	ec.AddString("method", method)
	b := ec.Marshal()
	// ## send req get call len
	code := sysCall(getRequestHeader(ContractMethodCallContractLen), string(b))
	if code != int32(SUCCESS) {
		return nil, ERROR
	}
	if valueLen == 0 {
		return nil, SUCCESS
	}

	// # get data
	// ## prepare param
	valueByte := make([]byte, valueLen)
	valuePtr = int32(uintptr(unsafe.Pointer(&valueByte[0])))
	ec.RemoveKey("value_ptr")
	ec.AddInt32("value_ptr", valuePtr)
	b = ec.Marshal()
	// ## send req get value
	code2 := sysCall(getRequestHeader(ContractMethodCallContract), string(b))
	if code2 != int32(SUCCESS) {
		return nil, ERROR
	}
	return valueByte, SUCCESS
}

func DeleteStateFromKey(key string) ResultCode {
	return DeleteState(key, "")
}

// SuccessResult record success data
func SuccessResult(msg string) {
	sysCall(getRequestHeader(ContractMethodSuccessResult), msg)
}

// SuccessResult record success data
func SuccessResultByte(msg []byte) {
	sysCall(getRequestHeader(ContractMethodSuccessResult), string(msg))
}

// ErrorResult record error msg
func ErrorResult(msg string) {
	sysCall(getRequestHeader(ContractMethodErrorResult), string(msg))
}

func GetCreatorOrgId() (string, ResultCode) {
	return stringArg(ContractParamCreatorOrgId)
}
func GetCreatorRole() (string, ResultCode) {
	return stringArg(ContractParamCreatorRole)
}
func GetCreatorPk() (string, ResultCode) {
	return stringArg(ContractParamCreatorPk)
}
func GetSenderOrgId() (string, ResultCode) {
	return stringArg(ContractParamSenderOrgId)
}
func GetSenderRole() (string, ResultCode) {
	return stringArg(ContractParamSenderRole)
}
func GetSenderPk() (string, ResultCode) {
	return stringArg(ContractParamSenderPk)
}
func GetBlockHeight() (string, ResultCode) {
	return stringArg(ContractParamBlockHeight)
}
func GetTxId() (string, ResultCode) {
	return stringArg(ContractParamTxId)
}
func getCtxPtr() int32 {
	if str, resultCode := stringArg(ContractParamContextPtr); resultCode != SUCCESS {
		LogMessage("failed to get ctx ptr")
		return 0
	} else {
		ptr, err := convert.StringToInt32(str) //string转int32
		if err != nil {
			LogMessage("get ptr err: " + err.Error())
		}
		return int32(ptr)
	}
}

func getArgsMap() error {
	if !argsFlag {
		argsMap = EasyUnmarshal(argsBytes)
		argsFlag = true
	}
	return nil
}
func stringArg(key string) (string, ResultCode) {
	result, code := Arg(key)
	return string(result), code
}
func Arg(key string) ([]byte, ResultCode) {
	err := getArgsMap()
	if err != nil {
		LogMessage("get Arg error:" + err.Error())
		return nil, ERROR
	}
	for _, v := range argsMap {
		if v.Key == key {
			return v.Value.([]byte), SUCCESS
		}
	}
	return nil, ERROR
}
func ArgString(key string) (string, ResultCode) {
	err := getArgsMap()
	if err != nil {
		LogMessage("get Arg error:" + err.Error())
		return "", ERROR
	}
	for _, v := range argsMap {
		if v.Key == key {
			return string(v.Value.([]byte)), SUCCESS
		}
	}
	return "", ERROR
}

func Args() []*EasyCodecItem {
	err := getArgsMap()
	if err != nil {
		LogMessage("get Args error:" + err.Error())
	}
	return argsMap
}

func (s *SimContextImpl) newIterator(startKey string, startField string, limitKey string, limitField string) (ResultSetKV, ResultCode) { //main.go中调用
	ec := NewEasyCodec()
	ec.AddString("start_key", startKey)
	ec.AddString("start_field", startField)
	ec.AddString("limit_key", limitKey)
	ec.AddString("limit_field", limitField)
	index, code := GetInt32FromChain(ec, ContractMethodKvIterator)
	return &ResultSetKvImpl{index}, code
}

func (s *SimContextImpl) NewIteratorWithField(key string, startField string, limitField string) (ResultSetKV, ResultCode) {
	return s.newIterator(key, startField, key, limitField)
}

// NewIterator
func (s *SimContextImpl) NewIterator(key string, limit string) (ResultSetKV, ResultCode) {
	return s.newIterator(key, "", limit, "")
}

func (s *SimContextImpl) NewIteratorPrefixWithKeyField(startKey string, startField string) (ResultSetKV, ResultCode) {
	ec := NewEasyCodec()
	ec.AddString("start_key", startKey)
	ec.AddString("start_field", startField)
	index, code := GetInt32FromChain(ec, ContractMethodKvPreIterator)
	return &ResultSetKvImpl{index}, code
}

func (s *SimContextImpl) NewIteratorPrefixWithKey(key string) (ResultSetKV, ResultCode) {
	return s.NewIteratorPrefixWithKeyField(key, "")
}

// ResultSet iterator query result KVdb
type ResultSetKvImpl struct { //为kv查询后的上下文
	index int32 // 链的句柄的index
}

func (r *ResultSetKvImpl) HasNext() bool {
	ec := NewEasyCodec()
	ec.AddInt32("rs_index", r.index)
	data, _ := GetInt32FromChain(ec, ContractMethodKvIteratorHasNext)
	return data != 0
}

func (r *ResultSetKvImpl) NextRow() (*EasyCodec, ResultCode) {
	ec := NewEasyCodec()
	ec.AddInt32("rs_index", r.index)
	bytes, code := GetBytesFromChain(ec, ContractMethodKvIteratorNextLen, ContractMethodKvIteratorNext)
	if code != SUCCESS {
		return nil, ERROR
	}
	ec = NewEasyCodecWithBytes(bytes)
	return ec, code
}

func (r *ResultSetKvImpl) Close() (bool, ResultCode) {
	ec := NewEasyCodec()
	ec.AddInt32("rs_index", r.index)
	data, code := GetInt32FromChain(ec, ContractMethodKvIteratorClose)
	return data != 0, code
}

func (r *ResultSetKvImpl) Next() (string, string, []byte, ResultCode) {
	ec, code := r.NextRow()
	if code != SUCCESS {
		return "", "", nil, ERROR
	}
	k, _ := ec.GetString("key")
	field, _ := ec.GetString("field")
	v, _ := ec.GetBytes("value")
	return k, field, v, code
}
