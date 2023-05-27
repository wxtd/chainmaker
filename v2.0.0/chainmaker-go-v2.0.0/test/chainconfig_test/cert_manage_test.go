/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package native_test

import (
	"fmt"
	"testing"

	apiPb "chainmaker.org/chainmaker/pb-go/v2/api"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"github.com/stretchr/testify/assert"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	native "chainmaker.org/chainmaker-go/test/chainconfig_test"
	"chainmaker.org/chainmaker-go/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 证书添加，个人添加自己的证书
func TestCertAdd(t *testing.T) {
	txId := utils.GetRandTxId()
	require.True(t, len(txId) > 0)
	fmt.Printf("\n============ send Tx [%s] ============\n", txId)

	// 添加证书 ../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt
	sk, member := native.GetUserSK(1)
	resp, err := native.UpdateSysRequest(sk, member, &native.InvokeContractMsg{TxId: txId, TxType: commonPb.TxType_INVOKE_CONTRACT,
		ChainId: CHAIN1, ContractName: syscontract.SystemContract_CERT_MANAGE.String(), MethodName: syscontract.CertManageFunction_CERT_ADD.String()})
	processResults(resp, err)
}

// 证书查询
func TestCertQuery(t *testing.T) {
	conn, err := native.InitGRPCConnect(isTls)
	require.NoError(t, err)
	client := apiPb.NewRpcNodeClient(conn)

	fmt.Println("============ get chain config by blockHeight============")
	// 构造Payload
	// 查询证书 ../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt
	pairs := []*commonPb.KeyValuePair{
		{
			Key:   "cert_hashes",
			Value: []byte("e77c9238c51e3446d942f94bd8803cc4f351254f8771f972146d7bfc6e0be7f4"),
		},
	}
	sk, member := native.GetUserSK(1)
	resp, err := native.QueryRequest(sk, member, &client, &native.InvokeContractMsg{TxType: commonPb.TxType_QUERY_CONTRACT, ChainId: CHAIN1,
		ContractName: syscontract.SystemContract_CERT_MANAGE.String(), MethodName: syscontract.CertManageFunction_CERTS_QUERY.String(), Pairs: pairs})
	processResults(resp, err)

	assert.Nil(t, err)
	c := &commonPb.CertInfos{}
	proto.Unmarshal(resp.ContractResult.Result, c)
	fmt.Printf("\n\n ========certs======== \n ")
	fmt.Println(c)
	assert.NotNil(t, c.CertInfos[0].Cert, "not found certs")
}

// 证书的删除（管理员操作）
func TestCertDelete(t *testing.T) {
	txId := utils.GetRandTxId()
	require.True(t, len(txId) > 0)
	fmt.Printf("\n============ send Tx [%s] ============\n", txId)

	// 构造Payload
	// 删除证书 ../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt
	pairs := []*commonPb.KeyValuePair{
		{
			Key:   "cert_hashes",
			Value: []byte("03725dc03b236f098153adea0fdf9a09dfe67fc8606a9ee1be7075c22e209a08"),
		},
	}
	sk, member := native.GetUserSK(1)
	resp, err := native.UpdateSysRequest(sk, member, &native.InvokeContractMsg{TxId: txId, TxType: commonPb.TxType_INVOKE_CONTRACT, ChainId: CHAIN1,
		ContractName: syscontract.SystemContract_CERT_MANAGE.String(), MethodName: syscontract.CertManageFunction_CERTS_DELETE.String(), Pairs: pairs})
	processResults(resp, err)
}

// 证书冻结
func TestCertFrozen(t *testing.T) {
	txId := utils.GetRandTxId()
	require.True(t, len(txId) > 0)
	// 构造Payload
	// 冻结证书 ../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt
	pairs := []*commonPb.KeyValuePair{
		{
			Key:   "certs",
			Value: []byte("-----BEGIN CERTIFICATE-----\nMIICijCCAi+gAwIBAgIDBS9vMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\nMTIwNzA2NTM0M1owgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMQ8wDQYDVQQLEwZjbGllbnQxLDAqBgNVBAMTI2NsaWVudDEuc2lnbi53eC1vcmcx\nLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE56xayRx0\n/a8KEXPxRfiSzYgJ/sE4tVeI/ZbjpiUX9m0TCJX7W/VHdm6WeJLOdCDuLLNvjGTy\nt8LLyqyubJI5AKN7MHkwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw\nKQYDVR0OBCIEIMjAiM2eMzlQ9HzV9ePW69rfUiRZVT2pDBOMqM4WVJSAMCsGA1Ud\nIwQkMCKAIDUkP3EcubfENS6TH3DFczH5dAnC2eD73+wcUF/bEIlnMAoGCCqGSM49\nBAMCA0kAMEYCIQCWUHL0xisjQoW+o6VV12pBXIRJgdeUeAu2EIjptSg2GAIhAIxK\nLXpHIBFxIkmWlxUaanCojPSZhzEbd+8LRrmhEO8n\n-----END CERTIFICATE-----\n"),
		},
	}

	sk, member := native.GetUserSK(1)
	resp, err := native.UpdateSysRequest(sk, member, &native.InvokeContractMsg{TxId: txId, TxType: commonPb.TxType_INVOKE_CONTRACT, ChainId: CHAIN1,
		ContractName: syscontract.SystemContract_CERT_MANAGE.String(), MethodName: syscontract.CertManageFunction_CERTS_FREEZE.String(), Pairs: pairs})
	processResults(resp, err)
}

// 证书解冻
func TestCertUnfrozen(t *testing.T) {
	txId := utils.GetRandTxId()
	require.True(t, len(txId) > 0)
	// 构造Payload
	// 解冻证书 ../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt
	pairs := []*commonPb.KeyValuePair{
		{
			Key:   "certs",
			Value: []byte("-----BEGIN CERTIFICATE-----\nMIICijCCAi+gAwIBAgIDBS9vMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\nMTIwNzA2NTM0M1owgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMQ8wDQYDVQQLEwZjbGllbnQxLDAqBgNVBAMTI2NsaWVudDEuc2lnbi53eC1vcmcx\nLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE56xayRx0\n/a8KEXPxRfiSzYgJ/sE4tVeI/ZbjpiUX9m0TCJX7W/VHdm6WeJLOdCDuLLNvjGTy\nt8LLyqyubJI5AKN7MHkwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw\nKQYDVR0OBCIEIMjAiM2eMzlQ9HzV9ePW69rfUiRZVT2pDBOMqM4WVJSAMCsGA1Ud\nIwQkMCKAIDUkP3EcubfENS6TH3DFczH5dAnC2eD73+wcUF/bEIlnMAoGCCqGSM49\nBAMCA0kAMEYCIQCWUHL0xisjQoW+o6VV12pBXIRJgdeUeAu2EIjptSg2GAIhAIxK\nLXpHIBFxIkmWlxUaanCojPSZhzEbd+8LRrmhEO8n\n-----END CERTIFICATE-----\n"),
		},
	}

	sk, member := native.GetUserSK(2)
	resp, err := native.UpdateSysRequest(sk, member, &native.InvokeContractMsg{TxId: txId, TxType: commonPb.TxType_INVOKE_CONTRACT, ChainId: CHAIN1,
		ContractName: syscontract.SystemContract_CERT_MANAGE.String(), MethodName: syscontract.CertManageFunction_CERTS_UNFREEZE.String(), Pairs: pairs})
	processResults(resp, err)
}

// 证书解冻
func TestCertUnfrozenWithCertHash(t *testing.T) {
	txId := utils.GetRandTxId()
	require.True(t, len(txId) > 0)
	// 构造Payload
	var pairs []*commonPb.KeyValuePair
	// 解冻证书 ../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt
	pairs = append(pairs, &commonPb.KeyValuePair{
		Key:   "cert_hashes",
		Value: []byte("e77c9238c51e3446d942f94bd8803cc4f351254f8771f972146d7bfc6e0be7f4,09ff34fafd2b97c8e9c7e05704b075d90cb7fee93cd2e4234e71cee6df0a88e6"),
	})

	sk, member := native.GetUserSK(2)
	resp, err := native.UpdateSysRequest(sk, member, &native.InvokeContractMsg{TxId: txId, TxType: commonPb.TxType_INVOKE_CONTRACT, ChainId: CHAIN1,
		ContractName: syscontract.SystemContract_CERT_MANAGE.String(), MethodName: syscontract.CertManageFunction_CERTS_UNFREEZE.String(), Pairs: pairs})
	processResults(resp, err)
}

// 证书吊销
func TestCertRevocation(t *testing.T) {
	txId := utils.GetRandTxId()
	require.True(t, len(txId) > 0)
	fmt.Println("============ get chain config by blockHeight in TestCertRevocation============")
	// 构造Payload
	var pairs []*commonPb.KeyValuePair
	// 吊销证书 ../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt
	pairs = append(pairs, &commonPb.KeyValuePair{
		Key: "cert_crl",
		// 多个就换行就行
		Value: []byte("-----BEGIN CRL-----\nMIIBXjCCAQMCAQEwCgYIKoZIzj0EAwIwgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQI\nEwdCZWlqaW5nMRAwDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNo\nYWlubWFrZXIub3JnMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4\nLW9yZzEuY2hhaW5tYWtlci5vcmcXDTIxMDcyMDEyMjYzMloXDTIxMDcyMDE2MjYz\nMlowFjAUAgMFL28XDTI0MDMyMzE1MDMwNVqgLzAtMCsGA1UdIwQkMCKAIDUkP3Ec\nubfENS6TH3DFczH5dAnC2eD73+wcUF/bEIlnMAoGCCqGSM49BAMCA0kAMEYCIQDy\nwvxZL30HRdyQYJzb1HsczH9xnh3iY+aW1ZbY46KX8AIhAPw8140++BTkBnlKBtAH\nPajXB4S3hsYlNv0RwV5Gfui4\n-----END CRL-----\n"),
	})

	sk, member := native.GetUserSK(1)
	resp, err := native.UpdateSysRequest(sk, member, &native.InvokeContractMsg{TxId: txId, TxType: commonPb.TxType_INVOKE_CONTRACT, ChainId: CHAIN1,
		ContractName: syscontract.SystemContract_CERT_MANAGE.String(), MethodName: syscontract.CertManageFunction_CERTS_REVOKE.String(), Pairs: pairs})
	processResults(resp, err)
}

func processResults(resp *commonPb.TxResponse, err error) {
	if err == nil {
		fmt.Printf("send tx resp: code:%d, msg:%s, payload:%+v\n", resp.Code, resp.Message, resp.ContractResult)
		return
	}
	if statusErr, ok := status.FromError(err); ok && statusErr.Code() == codes.DeadlineExceeded {
		fmt.Println("WARN: client.call err: deadline")
		return
	}
	fmt.Printf("ERROR: client.call err: %v\n", err)
}
