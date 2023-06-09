/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"chainmaker.org/chainmaker-go/logger"
	logger2 "chainmaker.org/chainmaker-go/logger"
	"chainmaker.org/chainmaker/common/v2/concurrentlru"
	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/stretchr/testify/require"
)

const (
	testChainId        = "chain1"
	testVersion        = "v1.0.0"
	testCertMemberType = "CERT"
	testHashType       = "SM3"

	testOrg1 = "org1"
	testOrg2 = "org2"
	testOrg3 = "org3"
	testOrg4 = "org4"
	testOrg5 = "org5"

	tempOrg1KeyFileName  = "org1.key"
	tempOrg1CertFileName = "org1.crt"

	testConsensusRole = protocol.Role("CONSENSUS")
	testAdminRole     = protocol.Role("ADMIN")
	testClientRole    = protocol.Role("CLIENT")

	testConsensusCN = "consensus1"
	testAdminCN     = "admin1"
	testClientCN    = "client1"

	testMsg = "Winter is coming."

	testCAOrg1 = `-----BEGIN CERTIFICATE-----
MIICQjCCAeegAwIBAgIIKrncHH+4b/swCgYIKoEcz1UBg3UwYjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxETAPBgNVBAoT
CG9yZy1yb290MQ0wCwYDVQQLEwRyb290MQ0wCwYDVQQDEwRyb290MB4XDTIxMDcw
NjAxNTkyNloXDTIzMDcwNjAxNTkyNlowXzELMAkGA1UEBhMCQ04xEDAOBgNVBAgT
B0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzExCzAJBgNV
BAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcxMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0D
QgAE9dFKJiTSUBfIDLZu7JWLVhT8mPMDvwAjGGkgIj5zDmAVWoTDkccbYuM+7Tn/
GgQK0f0BObK0TbNNuMRmuLUuhaOBiTCBhjAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T
AQH/BAUwAwEB/zApBgNVHQ4EIgQgsjfralJ8KKfnTEygIv2pBhXEhXymAGe+jGbE
AyTwmtAwKwYDVR0jBCQwIoAgt6lmGban5gHDAurtFWC1/Iv5waMl8w+Ld4d5+3fS
VUAwCwYDVR0RBAQwAoIAMAoGCCqBHM9VAYN1A0kAMEYCIQD0OcYuWyHhsKicunv4
ylpzmCqo49WkxzqILFi4TKqpEQIhAKUs+Gn6JxpDGSZX1b+DHqCel/+Vj4TFuhQP
hAPWXHIW
-----END CERTIFICATE-----
`
	testCAOrg2 = `-----BEGIN CERTIFICATE-----
MIICQjCCAeegAwIBAgIIL7YsID5bOCQwCgYIKoEcz1UBg3UwYjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxETAPBgNVBAoT
CG9yZy1yb290MQ0wCwYDVQQLEwRyb290MQ0wCwYDVQQDEwRyb290MB4XDTIxMDcw
NjAxNTkyNloXDTIzMDcwNjAxNTkyNlowXzELMAkGA1UEBhMCQ04xEDAOBgNVBAgT
B0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzIxCzAJBgNV
BAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcyMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0D
QgAE9zbC5g9R8kG08ld0YvCdRBiCj8aXPnGMBygQGBUZz0ZYHB4UKWGH61BUt7Ge
SR0wDY5f5U7gcfQwFeCupuMQKaOBiTCBhjAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T
AQH/BAUwAwEB/zApBgNVHQ4EIgQgH7iJPcW9+Bn00q6x76BbHe1O3vfGzNt/3ZJh
xGVpffwwKwYDVR0jBCQwIoAgt6lmGban5gHDAurtFWC1/Iv5waMl8w+Ld4d5+3fS
VUAwCwYDVR0RBAQwAoIAMAoGCCqBHM9VAYN1A0kAMEYCIQCVr7j8zTlXIXR7wZUG
t6aB3IfbEwLexZrwgobKoVWHCQIhANo5U3mc8Qp02d39L3J3hLAgoZ+cgQOeFJfL
5Q2+JVC8
-----END CERTIFICATE-----
`
	testCAOrg3 = `-----BEGIN CERTIFICATE-----
MIICQDCCAeegAwIBAgIIBwvsE5TejB4wCgYIKoEcz1UBg3UwYjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxETAPBgNVBAoT
CG9yZy1yb290MQ0wCwYDVQQLEwRyb290MQ0wCwYDVQQDEwRyb290MB4XDTIxMDcw
NjAxNTkyNloXDTIzMDcwNjAxNTkyNlowXzELMAkGA1UEBhMCQ04xEDAOBgNVBAgT
B0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzMxCzAJBgNV
BAsTAmNhMRAwDgYDVQQDEwdjYS1vcmczMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0D
QgAE618JNraSTiwh5uuyn96gu74AyY5c9D02yfZC8byiw+8GKTuOla5h6F8QdQrU
LWG5WsOPqUjE0vy59k3VBaVylaOBiTCBhjAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T
AQH/BAUwAwEB/zApBgNVHQ4EIgQgNR7+DQC5hfYiEUU/aM81BPdinm64TZdsmGHD
FZKeL9QwKwYDVR0jBCQwIoAgt6lmGban5gHDAurtFWC1/Iv5waMl8w+Ld4d5+3fS
VUAwCwYDVR0RBAQwAoIAMAoGCCqBHM9VAYN1A0cAMEQCIH55J+WejkPBYaC8vFYp
UoyAmSZ1hhwTawetC6bx8KIrAiBHEmwMDpYt23D/Jm8KOIj6TbqGEFNewCZt/oiY
d9qMUA==
-----END CERTIFICATE-----
`
	testCAOrg4 = `-----BEGIN CERTIFICATE-----
MIICQTCCAeegAwIBAgIIGeCcw3jEhs8wCgYIKoEcz1UBg3UwYjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxETAPBgNVBAoT
CG9yZy1yb290MQ0wCwYDVQQLEwRyb290MQ0wCwYDVQQDEwRyb290MB4XDTIxMDcw
NjAxNTkyNloXDTIzMDcwNjAxNTkyNlowXzELMAkGA1UEBhMCQ04xEDAOBgNVBAgT
B0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzQxCzAJBgNV
BAsTAmNhMRAwDgYDVQQDEwdjYS1vcmc0MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0D
QgAEddyootTCx8MjuvU/G06WhsSHaTQoVqx1KTNuzPp/2z8aO4lwrfe3lmHUXRq+
DeZsG7G41qpycFqQLsiiHxB+8qOBiTCBhjAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T
AQH/BAUwAwEB/zApBgNVHQ4EIgQgwyQkBxwULolBUwXXiI7holWi2KPMgVvjta0J
TFOaOj8wKwYDVR0jBCQwIoAgt6lmGban5gHDAurtFWC1/Iv5waMl8w+Ld4d5+3fS
VUAwCwYDVR0RBAQwAoIAMAoGCCqBHM9VAYN1A0gAMEUCIQDpz3KDL3FwDw5kcg+3
Kgxv5JQDgywXtBiobwcRs+jbqAIgNWqzPOyMYgjtVSKLkAq9RshOZATHp+pa5pOC
+obIQlc=
-----END CERTIFICATE-----
`
	testTrustMember1 = `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIIKVbkVBlA0XYwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQwNDda
Fw0yMzA4MDUwMzQwNDdaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4xMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAELf71
DQTS9zpzs3nDDdt6ncocPHrlqdpZvobToTNPeYmrIFBuahrokQZ14CvxZP632KJk
ohAlGfAfoxsdciuIiaN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCAuvneJ
0X1P7/K9yZRF+I0VEEWrFWTmqkq4In9l45GAFTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJUmhAHycQXCV68HnQvF761kE5157fXoQB6huFKBj1ySAiEA87/G
VF6kotuIP24ujAzANvkoZJeOhpk1hVS2xdIZ86s=
-----END CERTIFICATE-----
`
	testTrustMember2 = `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIILBJts5OBl+8wCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQyMDVa
Fw0yMzA4MDUwMzQyMDVaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4yMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAExoxt
S//rEqnhj6/ZNxxmfFY767XyeZrbrxewTtYqLZJYwOik3CsVhSsrelgAdsBOG4Pe
o7eCet9lxpq2NM/XBKN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCDbJtdv
0Krwa+vHyNB2urb8XC54OFy8oAwfvTbKc9l8iTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJZqy5zdbQ3ZfAJci7QcKLkMNXoqz2VHxH0QXz26uDvUAiEAgmIc
Ds20ILx7wy349jvs8s4Rc1P4hJZQdfkxdI2GhXU=
-----END CERTIFICATE-----
`
)

var testChainConfig = &config.ChainConfig{
	ChainId:    testChainId,
	Version:    testVersion,
	MemberType: testCertMemberType,
	Sequence:   0,
	Crypto: &config.CryptoConfig{
		Hash: testHashType,
	},
	Block: nil,
	Core:  nil,
	Consensus: &config.ConsensusConfig{
		Type: 0,
		Nodes: []*config.OrgConfig{{
			OrgId:  testOrg1,
			NodeId: nil,
		}, {
			OrgId:  testOrg2,
			NodeId: nil,
		}, {
			OrgId:  testOrg3,
			NodeId: nil,
		}, {
			OrgId:  testOrg4,
			NodeId: nil,
		},
		},
		ExtConfig: nil,
	},
	TrustRoots: []*config.TrustRootConfig{
		{
			OrgId: testOrg1,
			Root:  []string{testCAOrg1},
		},
		{
			OrgId: testOrg2,
			Root:  []string{testCAOrg2},
		},
		{
			OrgId: testOrg3,
			Root:  []string{testCAOrg3},
		},
		{
			OrgId: testOrg4,
			Root:  []string{testCAOrg4},
		},
	},
	TrustMembers: []*config.TrustMemberConfig{
		{
			OrgId:      testOrg5,
			Role:       "admin",
			MemberInfo: testTrustMember1,
		},
		{
			OrgId:      testOrg5,
			Role:       "admin",
			MemberInfo: testTrustMember2,
		},
	},
}

type testCertInfo struct {
	cert string
	sk   string
}

var testConsensusSignOrg1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICPzCCAeagAwIBAgIINI6eOha7f6gwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzExCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcxMB4XDTIxMDcwNjAy
MDMwMloXDTIzMDcwNjAyMDMwMlowaTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzExEjAQBgNVBAsT
CWNvbnNlbnN1czETMBEGA1UEAxMKY29uc2Vuc3VzMTBZMBMGByqGSM49AgEGCCqB
HM9VAYItA0IABCJXY8cI0MHTzIIJo9TgMRKhGIliKuJjNJEvBvdMG3Ew/V87vy7q
2TBRMqJeVtPRP2a2vc/jXTICOWTr4RK4eHqjgYEwfzAOBgNVHQ8BAf8EBAMCBsAw
KQYDVR0OBCIEIEychptOEN+2695oymyT8Iy3g8Kv3l8H+5XwyJOef5+bMCsGA1Ud
IwQkMCKAILI362pSfCin50xMoCL9qQYVxIV8pgBnvoxmxAMk8JrQMBUGA1UdEQQO
MAyCCmNvbnNlbnN1czEwCgYIKoEcz1UBg3UDRwAwRAIgdxPwLEjwhNBmPRpjj6GA
X7fy5bEHffwmPYRMnEVUetACIGE2woYmGrAbfxYQ6K9zGLTvkaZhY7tCX4cow8Wy
IEbL
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgGQ35DUKL2CUNX1Xs
7e40KoBIneCJvzd6/Ko3qY7jQc+gCgYIKoEcz1UBgi2hRANCAAQiV2PHCNDB08yC
CaPU4DESoRiJYiriYzSRLwb3TBtxMP1fO78u6tkwUTKiXlbT0T9mtr3P410yAjlk
6+ESuHh6
-----END PRIVATE KEY-----
`,
}
var testConsensusTlsOrg1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICYTCCAgagAwIBAgIIDVVwVu/XE4owCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzExCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcxMB4XDTIxMDcwNjAy
NTYwM1oXDTIzMDcwNjAyNTYwM1owaTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzExEjAQBgNVBAsT
CWNvbnNlbnN1czETMBEGA1UEAxMKY29uc2Vuc3VzMTBZMBMGByqGSM49AgEGCCqB
HM9VAYItA0IABDehr437Fiz2+p6uwq8wgC5UoaA9atkaOdd+HhmUt3SfMZfKnGk9
sJ5N8+i00Vz8zA7MeCZfS6jfbExdT0EopFKjgaEwgZ4wDgYDVR0PAQH/BAQDAgP4
MB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATApBgNVHQ4EIgQgTzu3yvoP
jxpdPYRkBWztBa2Pn0QAZD/NtS5YU2AEdk0wKwYDVR0jBCQwIoAgsjfralJ8KKfn
TEygIv2pBhXEhXymAGe+jGbEAyTwmtAwFQYDVR0RBA4wDIIKY29uc2Vuc3VzMTAK
BggqgRzPVQGDdQNJADBGAiEAh7Jg6cLNl5JtOu4yhDRb1OGhsd0K25IUlNk3z6qv
258CIQCL9To7rVk4wNcUFZ8m858tN4VZicgmmJDiDRowo4FDwA==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg5QriJPiflpuB/S1d
uIgA8OFj5+n+VZK6mEgI7tV39+ygCgYIKoEcz1UBgi2hRANCAAQ3oa+N+xYs9vqe
rsKvMIAuVKGgPWrZGjnXfh4ZlLd0nzGXypxpPbCeTfPotNFc/MwOzHgmX0uo32xM
XU9BKKRS
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICQDCCAeagAwIBAgIIJfZ6/YWW704wCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzIxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcyMB4XDTIxMDcwNjAy
MjcyMVoXDTIzMDcwNjAyMjcyMVowaTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzIxEjAQBgNVBAsT
CWNvbnNlbnN1czETMBEGA1UEAxMKY29uc2Vuc3VzMTBZMBMGByqGSM49AgEGCCqB
HM9VAYItA0IABF7M1FyGnBPKqT1PPlUKtWJ22LKax2tj3w5Fp02iHwPwsaS8wXM9
MEy0pJ+8RchO4WOCXEJFML32d19Q4lQ39AGjgYEwfzAOBgNVHQ8BAf8EBAMCBsAw
KQYDVR0OBCIEIP5eFhWvosUDYgVuGrwb14JmHI6mWN9WJ0PKq3xof+bcMCsGA1Ud
IwQkMCKAIB+4iT3FvfgZ9NKuse+gWx3tTt73xszbf92SYcRlaX38MBUGA1UdEQQO
MAyCCmNvbnNlbnN1czEwCgYIKoEcz1UBg3UDSAAwRQIgDWLqzW1OeZBsBm4GF4fe
2Mde5DneE6NcwKMcnkisxeQCIQCDWkVVkTn58avjWgBxk0DYeuX3vuB3ATxq7bw1
W3VLGA==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg9oWqW25x6XwVgiFc
U9UXN1B3LLUpOeEYYMfwts7Qn7mgCgYIKoEcz1UBgi2hRANCAARezNRchpwTyqk9
Tz5VCrVidtiymsdrY98ORadNoh8D8LGkvMFzPTBMtKSfvEXITuFjglxCRTC99ndf
UOJUN/QB
-----END PRIVATE KEY-----
`,
}
var testConsensusTlsOrg2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICYTCCAgagAwIBAgIIFse+TqTc/PgwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzIxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcyMB4XDTIxMDcwNjAy
NTk1NloXDTIzMDcwNjAyNTk1NlowaTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzIxEjAQBgNVBAsT
CWNvbnNlbnN1czETMBEGA1UEAxMKY29uc2Vuc3VzMTBZMBMGByqGSM49AgEGCCqB
HM9VAYItA0IABB814d5OtAC1Edevci6J7AS4W8s9mMox2uVJzM8YNEHbNKj0cVH3
wfPAvCycBFcXrvbcMEZ608ur8+5LkM7to26jgaEwgZ4wDgYDVR0PAQH/BAQDAgP4
MB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATApBgNVHQ4EIgQg4LsPC5s0
s/Mo5frmzr1ctwJP6uQa+sk24bSzs8bFlP4wKwYDVR0jBCQwIoAgH7iJPcW9+Bn0
0q6x76BbHe1O3vfGzNt/3ZJhxGVpffwwFQYDVR0RBA4wDIIKY29uc2Vuc3VzMTAK
BggqgRzPVQGDdQNJADBGAiEA+lzq2ghorbkxjcX1MHDhQguuKM8hkaS2H6EOHqCK
+98CIQDBgXpSTUS9PDz1aLqrYYcEzkiuohAFdKB6d2MICadl0Q==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgbY8C0eLqJsTr25zS
KKb/JOqIadvvQ2ehLs24P7KjEX2gCgYIKoEcz1UBgi2hRANCAAQfNeHeTrQAtRHX
r3IuiewEuFvLPZjKMdrlSczPGDRB2zSo9HFR98HzwLwsnARXF6723DBGetPLq/Pu
S5DO7aNu
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg3 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICQTCCAeagAwIBAgIIGQAHGbAzkSgwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzMxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmczMB4XDTIxMDcwNjAy
MzQ0NloXDTIzMDcwNjAyMzQ0NlowaTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzMxEjAQBgNVBAsT
CWNvbnNlbnN1czETMBEGA1UEAxMKY29uc2Vuc3VzMTBZMBMGByqGSM49AgEGCCqB
HM9VAYItA0IABAhm8cO9Fkvq68ynUNx9ZfqInVFn1UujiCBuIrw8g60hUOxkk7I5
lqiLSZGFvVZDQvslUO7zL+oORQtfxAMY1zyjgYEwfzAOBgNVHQ8BAf8EBAMCBsAw
KQYDVR0OBCIEIMljO9MWe8SzQOgVP1XaR9TBFYJWcsO6qs5bJustIxZNMCsGA1Ud
IwQkMCKAIDUe/g0AuYX2IhFFP2jPNQT3Yp5uuE2XbJhhwxWSni/UMBUGA1UdEQQO
MAyCCmNvbnNlbnN1czEwCgYIKoEcz1UBg3UDSQAwRgIhAPkNQbN73aYNaKUt6xCI
u3RbLp+bZL2+gF2psN4F6CInAiEAvoY1i8ZLi7ki57zL6bbqrg4Bb4RvkRFEjgYJ
LcUXkyw=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg6Muc8+uaA6Kfx1DV
7CL0lhmZuEthu7CtVWBY5WZXk7OgCgYIKoEcz1UBgi2hRANCAAQIZvHDvRZL6uvM
p1DcfWX6iJ1RZ9VLo4ggbiK8PIOtIVDsZJOyOZaoi0mRhb1WQ0L7JVDu8y/qDkUL
X8QDGNc8
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg4 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICQDCCAeagAwIBAgIIJnerFjJ6aokwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzQxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmc0MB4XDTIxMDcwNjAy
NDQ0OFoXDTIzMDcwNjAyNDQ0OFowaTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzQxEjAQBgNVBAsT
CWNvbnNlbnN1czETMBEGA1UEAxMKY29uc2Vuc3VzMTBZMBMGByqGSM49AgEGCCqB
HM9VAYItA0IABPA027H39SVdtTcuWX16Cv5dck466gjE+FGqTSn6zxaRIVE0YP2S
6lbuI8WQoyBHrVKHD9GPbJsrN10POZF7hRSjgYEwfzAOBgNVHQ8BAf8EBAMCBsAw
KQYDVR0OBCIEID+f4xyLZhu4yGIR4cc/CGSlTXChz5/ZBvsEnhlAbBWCMCsGA1Ud
IwQkMCKAIMMkJAccFC6JQVMF14iO4aJVotijzIFb47WtCUxTmjo/MBUGA1UdEQQO
MAyCCmNvbnNlbnN1czEwCgYIKoEcz1UBg3UDSAAwRQIhAPSy1q+1Baq5mIgzEh7h
NVwT0+Z61lQc6G2D55kQIYd8AiB7LKjK167hZ1KmUTQgXsd5fjtHl62Y+GAhsT6r
/9yI6Q==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgTn4bymGnJobO/S1L
zccK3rqhiSZe1qLUyE2mOcPzFkegCgYIKoEcz1UBgi2hRANCAATwNNux9/UlXbU3
Lll9egr+XXJOOuoIxPhRqk0p+s8WkSFRNGD9kupW7iPFkKMgR61Shw/Rj2ybKzdd
DzmRe4UU
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICLDCCAdOgAwIBAgIIBqeveockqiMwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzExCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcxMB4XDTIxMDcwNjAy
MDYxOVoXDTIzMDcwNjAyMDYxOVowYTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzExDjAMBgNVBAsT
BWFkbWluMQ8wDQYDVQQDEwZhZG1pbjEwWTATBgcqhkjOPQIBBggqgRzPVQGCLQNC
AASYaLUmquus067fH/dO7tVcZacmc+xDowea7dcruE9AhKhgUO8wpw2wq67uAlQo
wqOkRf7rNTPV/ZMna6VtQhyoo3cwdTAOBgNVHQ8BAf8EBAMCBsAwKQYDVR0OBCIE
ILF+asxjFbysg6s5QpBHFTLZcKnEPmfs5aZqaB/UhZXMMCsGA1UdIwQkMCKAILI3
62pSfCin50xMoCL9qQYVxIV8pgBnvoxmxAMk8JrQMAsGA1UdEQQEMAKCADAKBggq
gRzPVQGDdQNHADBEAiA6l+ITYhoO1+VIt0LBtpqhN4FBP6t3wD/UIpQCO7Xe/AIg
ZHlsnoaGqokzktPdPrOuzcl/1O+bRzH1Lxc5bpF55DU=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgfIiR3EfbCgDeF1o4
C+aqLIn1NDA9+L4At2u8TtYA70GgCgYIKoEcz1UBgi2hRANCAASYaLUmquus067f
H/dO7tVcZacmc+xDowea7dcruE9AhKhgUO8wpw2wq67uAlQowqOkRf7rNTPV/ZMn
a6VtQhyo
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICLTCCAdOgAwIBAgIIBzj8/aB96dYwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzIxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcyMB4XDTIxMDcwNjAy
MzAyMFoXDTIzMDcwNjAyMzAyMFowYTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzIxDjAMBgNVBAsT
BWFkbWluMQ8wDQYDVQQDEwZhZG1pbjEwWTATBgcqhkjOPQIBBggqgRzPVQGCLQNC
AARj4P/Uj4KgzIA3clVovGjkhpUiXuDRrhJg5koweCiikRx/x1Ip6xnCXmgwaZ+S
tErC+QDtfj5U0Ri37Ubw3I3Xo3cwdTAOBgNVHQ8BAf8EBAMCBsAwKQYDVR0OBCIE
ILM4AdYo3xeYEYjqTmmlUYtZO4zlZ5ScyWbY6+eRHC4qMCsGA1UdIwQkMCKAIB+4
iT3FvfgZ9NKuse+gWx3tTt73xszbf92SYcRlaX38MAsGA1UdEQQEMAKCADAKBggq
gRzPVQGDdQNIADBFAiEA7pLqU3LjMJpG9BlW1aRL4iyIbZsoqQJlcCZXBKV+th4C
IF557eyvfbyQU57NEGIf/hyJi8b7Q9YuEJZ/TiMUZrm4
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgjVikTkBGM1EYy+1W
E1Z3IaqQo6TzobkphWgn6E4UU9igCgYIKoEcz1UBgi2hRANCAARj4P/Uj4KgzIA3
clVovGjkhpUiXuDRrhJg5koweCiikRx/x1Ip6xnCXmgwaZ+StErC+QDtfj5U0Ri3
7Ubw3I3X
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg3 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICLTCCAdOgAwIBAgIIG3ugGtC1KfEwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzMxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmczMB4XDTIxMDcwNjAy
MzczMFoXDTIzMDcwNjAyMzczMFowYTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzMxDjAMBgNVBAsT
BWFkbWluMQ8wDQYDVQQDEwZhZG1pbjEwWTATBgcqhkjOPQIBBggqgRzPVQGCLQNC
AASam2S6khjWafBL1k5/2JM1A9ADEwE8Ya7/A+6G4w6i2quZR4iZLSb4ygkUsMaz
8r9P2IlJHPfEIFdh9pK+q3eOo3cwdTAOBgNVHQ8BAf8EBAMCBsAwKQYDVR0OBCIE
ILg7hlLNyQPe4tYVs1UyHXM8aVjfYiLIaN4bjb2pMiIkMCsGA1UdIwQkMCKAIDUe
/g0AuYX2IhFFP2jPNQT3Yp5uuE2XbJhhwxWSni/UMAsGA1UdEQQEMAKCADAKBggq
gRzPVQGDdQNIADBFAiBqBLDaDBNfQxKCeOq7jKN+itOZx4jUIH2cd1vavMnjCAIh
ALPiK+cNTkF1N6qqNmwLya2A3lDMLgqEGUZnAi0yPrl/
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgZYJGp50wQxcdAxsb
049RpxjzjhoeDt5jKvPHfkk/DvigCgYIKoEcz1UBgi2hRANCAASam2S6khjWafBL
1k5/2JM1A9ADEwE8Ya7/A+6G4w6i2quZR4iZLSb4ygkUsMaz8r9P2IlJHPfEIFdh
9pK+q3eO
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg4 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICLTCCAdOgAwIBAgIIAzdWiZnaig8wCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzQxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmc0MB4XDTIxMDcwNjAy
NDczNloXDTIzMDcwNjAyNDczNlowYTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzQxDjAMBgNVBAsT
BWFkbWluMQ8wDQYDVQQDEwZhZG1pbjEwWTATBgcqhkjOPQIBBggqgRzPVQGCLQNC
AASm29iL1zHNd8IPKdG59cFedomktTdbXJEx9ALixTJgxGAZ66IOegAB1ytmbww4
ro/r6HLK43JUMwt8u+syNtHJo3cwdTAOBgNVHQ8BAf8EBAMCBsAwKQYDVR0OBCIE
IOwh1g9PZVInZlFSYWgqIMeE+931XgKAzW69qpkUvgKTMCsGA1UdIwQkMCKAIMMk
JAccFC6JQVMF14iO4aJVotijzIFb47WtCUxTmjo/MAsGA1UdEQQEMAKCADAKBggq
gRzPVQGDdQNIADBFAiEA701vzfjAlGzXrfKphzotED6WnK2nvh7KtUnAkAsKF78C
IBpQWAARrUyv4qfCv0zJAdxvu9uLp5xdr0H1oGGM2Kii
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgIzSV5H7LpaCsIdmr
rmJ5uuQWaIumtJRTsgfdWdlsMfegCgYIKoEcz1UBgi2hRANCAASm29iL1zHNd8IP
KdG59cFedomktTdbXJEx9ALixTJgxGAZ66IOegAB1ytmbww4ro/r6HLK43JUMwt8
u+syNtHJ
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICLjCCAdWgAwIBAgIIDvSW8zrwoz0wCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzExCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcxMB4XDTIxMDcwNjAy
MDg1MloXDTIzMDcwNjAyMDg1MlowYzELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzExDzANBgNVBAsT
BmNsaWVudDEQMA4GA1UEAxMHY2xpZW50MTBZMBMGByqGSM49AgEGCCqBHM9VAYIt
A0IABFM8GhZLvKL/mmDyKSqHbnmiyR6SBTIYawwajGeRCmUx06codlj3dRGc+SnH
eZOFj1YO567hJ7eqDRMC+Aj4EwajdzB1MA4GA1UdDwEB/wQEAwIGwDApBgNVHQ4E
IgQgI/qBIADaod7yDMkUMEWCdBJjcd2EtMnQcfWPpaVg9H4wKwYDVR0jBCQwIoAg
sjfralJ8KKfnTEygIv2pBhXEhXymAGe+jGbEAyTwmtAwCwYDVR0RBAQwAoIAMAoG
CCqBHM9VAYN1A0cAMEQCIGSrE8b8olUsn1cBrtSWGaW8ERxwm6hpbCQpx+/V5YhX
AiBBYOpg4ypidD9OuQAYax+qx8MHn6TNlKZIH4YiDjAsZw==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgCbczXL67OSt/Q5sK
8FDV9yGQVtMjYCfMbKdp2iPOFDCgCgYIKoEcz1UBgi2hRANCAARTPBoWS7yi/5pg
8ikqh255oskekgUyGGsMGoxnkQplMdOnKHZY93URnPkpx3mThY9WDueu4Se3qg0T
AvgI+BMG
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICLzCCAdWgAwIBAgIILeOrNzdAOHEwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzIxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmcyMB4XDTIxMDcwNjAy
MzI0MloXDTIzMDcwNjAyMzI0MlowYzELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzIxDzANBgNVBAsT
BmNsaWVudDEQMA4GA1UEAxMHY2xpZW50MTBZMBMGByqGSM49AgEGCCqBHM9VAYIt
A0IABK39h91yPfKcuAcOA0dWer7oyEXOV42Ro2y+OpoO5znErRmdQw0+dbAztVln
LuSTBZSGMRi4gF3P5fX28zEbxhmjdzB1MA4GA1UdDwEB/wQEAwIGwDApBgNVHQ4E
IgQgbUjHYOMOFbd7FoCAPfFR4hstD0gVIRdBTaMJbu1XBaUwKwYDVR0jBCQwIoAg
H7iJPcW9+Bn00q6x76BbHe1O3vfGzNt/3ZJhxGVpffwwCwYDVR0RBAQwAoIAMAoG
CCqBHM9VAYN1A0gAMEUCIQDUgFifSZtvKe45kp8OLt64sJRLH5gyRVxfG2qZu9IY
rgIgP1zA8KwQcl1sOwhhsU+wflJ3sJldVpMDZ9tQB3DHG1E=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg8EpA2IQjp4m85h+i
JsFvuaizkLSsCUcucmjGWOys1yugCgYIKoEcz1UBgi2hRANCAASt/Yfdcj3ynLgH
DgNHVnq+6MhFzleNkaNsvjqaDuc5xK0ZnUMNPnWwM7VZZy7kkwWUhjEYuIBdz+X1
9vMxG8YZ
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg3 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICMDCCAdWgAwIBAgIIBd1NN3fAoPkwCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzMxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmczMB4XDTIxMDcwNjAy
NDA1N1oXDTIzMDcwNjAyNDA1N1owYzELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzMxDzANBgNVBAsT
BmNsaWVudDEQMA4GA1UEAxMHY2xpZW50MTBZMBMGByqGSM49AgEGCCqBHM9VAYIt
A0IABEbphb4ydhdIKtj4zpmza/CLBkGNvaMKIm2P/HuVwhiNcDyjhPTpUCgcRjUy
hmMHASivuvy4VH2AbGtXeK2WTJGjdzB1MA4GA1UdDwEB/wQEAwIGwDApBgNVHQ4E
IgQgTXtr18VKkdMGK1jnoIBUbuMgmLyPrGTQTPvV4e5qPiMwKwYDVR0jBCQwIoAg
NR7+DQC5hfYiEUU/aM81BPdinm64TZdsmGHDFZKeL9QwCwYDVR0RBAQwAoIAMAoG
CCqBHM9VAYN1A0kAMEYCIQDA/wsi0vH9zFN1FHaXCmxLiuB3blZFG1x3+CkMKfQG
oQIhAOrjnfn5xlA1qL3NAwWYVGo9StbEZPe+LCDSKp25zsjg
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg33/j+evTYIdhmXOl
H3Kbb99Nstg7iUfPbjT9VhrLHsugCgYIKoEcz1UBgi2hRANCAARG6YW+MnYXSCrY
+M6Zs2vwiwZBjb2jCiJtj/x7lcIYjXA8o4T06VAoHEY1MoZjBwEor7r8uFR9gGxr
V3itlkyR
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg4 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICMDCCAdWgAwIBAgIIMc3Gy3tDmt4wCgYIKoEcz1UBg3UwXzELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzQxCzAJBgNVBAsTAmNhMRAwDgYDVQQDEwdjYS1vcmc0MB4XDTIxMDcwNjAy
NTAwMloXDTIzMDcwNjAyNTAwMlowYzELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0Jl
aWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoTBG9yZzQxDzANBgNVBAsT
BmNsaWVudDEQMA4GA1UEAxMHY2xpZW50MTBZMBMGByqGSM49AgEGCCqBHM9VAYIt
A0IABGNBJVpD6mORcL7vMvVbwABNPFCOvxr3mknFYBGIL/gHEOfCrUsdmkz5BjkA
TS2Zm+d/9IBHxorMQ6m/Ch6q13KjdzB1MA4GA1UdDwEB/wQEAwIGwDApBgNVHQ4E
IgQgmgfNeIifjEk8oC/9JVqBlo8dqCJgtMsEJwFK+nKGa0UwKwYDVR0jBCQwIoAg
wyQkBxwULolBUwXXiI7holWi2KPMgVvjta0JTFOaOj8wCwYDVR0RBAQwAoIAMAoG
CCqBHM9VAYN1A0kAMEYCIQDqQ7A9LoV4FA646ml2U+takpNreRtVKanFSTmvxGjV
fAIhANMLIfdKgzBXmQkkRkEByBGNXxLkWSsVX6Catu59REMA
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgoWZEODytNFVqGuD8
wXcA1N5egl65no8ub5lWbDtpSX6gCgYIKoEcz1UBgi2hRANCAARjQSVaQ+pjkXC+
7zL1W8AATTxQjr8a95pJxWARiC/4BxDnwq1LHZpM+QY5AE0tmZvnf/SAR8aKzEOp
vwoeqtdy
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg5 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICOjCCAeGgAwIBAgIIGN/iRBNkA0kwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUxMjE1NTla
Fw0yMzA4MDUxMjE1NTlaMGkxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MRIwEAYDVQQLEwljb25z
ZW5zdXMxEzARBgNVBAMTCmNvbnNlbnN1czEwWTATBgcqhkjOPQIBBggqgRzPVQGC
LQNCAATrDM8W9PU/9idSEGLXbCneUqlrY5ExNWShWg+1Qy8p1rDtwpLFTEuDR6sf
kQV8T9i1zeXefyS066zJZnhBpyJWo4GBMH8wDgYDVR0PAQH/BAQDAgbAMCkGA1Ud
DgQiBCDmj9z0hrOaUVRwkG6YlPxXarHRD37KGLU4YdOYre0aATArBgNVHSMEJDAi
gCCrqUGCeuslhFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTAVBgNVHREEDjAMggpj
b25zZW5zdXMxMAoGCCqBHM9VAYN1A0cAMEQCIEKFF/F682Ok2SO1dMUsVpKWmIBa
DagEEDWacKJ/07bAAiBevhEXM+6cqblRcTqLPQMNG6+Xz/gwcvHww8k9GspwRA==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgkycHiKroE0z0AQmT
zX18Q2ia2YpytlMxzHJ1hxKUn8igCgYIKoEcz1UBgi2hRANCAATrDM8W9PU/9idS
EGLXbCneUqlrY5ExNWShWg+1Qy8p1rDtwpLFTEuDR6sfkQV8T9i1zeXefyS066zJ
ZnhBpyJW
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg5 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIIJb8IwGdGzDMwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUxMjIyMjda
Fw0yMzA4MDUxMjIyMjdaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4zMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEPuP1
YyIVkKYfHQa2iR6A2E5MnnQYftIu6UhKRZI4EDT/DDs4l+2ksfTf4YeJUQqailwe
QESUFyyhXPWWKU0yDaN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCCE6BKh
MBVt9KFx7LmE3wXkILhbdTX07ZY8dKTDglFCIjArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAIudP9N2PbqWyOFrJKUwW5qO51hQciQsKKyLY8YTafsRAiEAmND5
BpWsfd537YspBgQRBDg5ztVRc68wp3C4AdqWc5Q=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgvCLtqX1Ser0F8H+2
1lCKQw54umUBz97DxcCdR6/ur2CgCgYIKoEcz1UBgi2hRANCAAQ+4/VjIhWQph8d
BraJHoDYTkyedBh+0i7pSEpFkjgQNP8MOziX7aSx9N/hh4lRCpqKXB5ARJQXLKFc
9ZYpTTIN
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg5 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKjCCAdCgAwIBAgIIKi+Lqj7RJ50wCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUxMjIwNDVa
Fw0yMzA4MDUxMjIwNDVaMGMxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ8wDQYDVQQLEwZjbGll
bnQxEDAOBgNVBAMTB2NsaWVudDEwWTATBgcqhkjOPQIBBggqgRzPVQGCLQNCAARU
XOhCeTsKsPBYt2xEyxYGSBY6xuNwcj0ppqcMTnH9J6javljoxDpKpNF2tcFK6CA3
/Z9j/APE7s5vkZK2W7Czo3cwdTAOBgNVHQ8BAf8EBAMCBsAwKQYDVR0OBCIEIAcw
XrkZCq8G6XkdUKqiNeQfijHC+VLKXfmtdKk3ADp5MCsGA1UdIwQkMCKAIKupQYJ6
6yWEU2vDnoJeUjWTAvmtys/FwZDpkJdt6GgxMAsGA1UdEQQEMAKCADAKBggqgRzP
VQGDdQNIADBFAiAy+gAGzbZcGIP17iKzyYBpu2qIEs9CXaM45AUelLb4QwIhAKA3
uZFq5Yw+M+1RCZm1JWYEqICxws5LW4I5vxoEM7F/
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgzwNxhFP8qFtIv2WZ
n4X2TFihYUnESnfnQ7M2kPkI//CgCgYIKoEcz1UBgi2hRANCAARUXOhCeTsKsPBY
t2xEyxYGSBY6xuNwcj0ppqcMTnH9J6javljoxDpKpNF2tcFK6CA3/Z9j/APE7s5v
kZK2W7Cz
-----END PRIVATE KEY-----
`,
}

var testTrustMemberAdmin1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIIKVbkVBlA0XYwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQwNDda
Fw0yMzA4MDUwMzQwNDdaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4xMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAELf71
DQTS9zpzs3nDDdt6ncocPHrlqdpZvobToTNPeYmrIFBuahrokQZ14CvxZP632KJk
ohAlGfAfoxsdciuIiaN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCAuvneJ
0X1P7/K9yZRF+I0VEEWrFWTmqkq4In9l45GAFTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJUmhAHycQXCV68HnQvF761kE5157fXoQB6huFKBj1ySAiEA87/G
VF6kotuIP24ujAzANvkoZJeOhpk1hVS2xdIZ86s=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgoSPsIukE1OdHTtwd
AJm1b4QjzOVv+B6N+M9rCb6OELSgCgYIKoEcz1UBgi2hRANCAAQt/vUNBNL3OnOz
ecMN23qdyhw8euWp2lm+htOhM095iasgUG5qGuiRBnXgK/Fk/rfYomSiECUZ8B+j
Gx1yK4iJ
-----END PRIVATE KEY-----
`,
}

var testTrustMemberAdmin2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIILBJts5OBl+8wCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQyMDVa
Fw0yMzA4MDUwMzQyMDVaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4yMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAExoxt
S//rEqnhj6/ZNxxmfFY767XyeZrbrxewTtYqLZJYwOik3CsVhSsrelgAdsBOG4Pe
o7eCet9lxpq2NM/XBKN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCDbJtdv
0Krwa+vHyNB2urb8XC54OFy8oAwfvTbKc9l8iTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJZqy5zdbQ3ZfAJci7QcKLkMNXoqz2VHxH0QXz26uDvUAiEAgmIc
Ds20ILx7wy349jvs8s4Rc1P4hJZQdfkxdI2GhXU=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgReLJhNH+Hn8RNyYO
OnXvLVHEDt500mN060ym0gl52N6gCgYIKoEcz1UBgi2hRANCAATGjG1L/+sSqeGP
r9k3HGZ8VjvrtfJ5mtuvF7BO1iotkljA6KTcKxWFKyt6WAB2wE4bg96jt4J632XG
mrY0z9cE
-----END PRIVATE KEY-----
`,
}

func createTempDirWithCleanFunc() (string, func(), error) {
	var td = filepath.Join("./temp")
	err := os.MkdirAll(td, os.ModePerm)
	if err != nil {
		return "", nil, err
	}
	var cleanFunc = func() {
		_ = os.RemoveAll(td)
		_ = os.RemoveAll(filepath.Join("./default.log"))
		now := time.Now()
		_ = os.RemoveAll(filepath.Join("./default.log." + now.Format("2006010215")))
		now = now.Add(-2 * time.Hour)
		_ = os.RemoveAll(filepath.Join("./default.log." + now.Format("2006010215")))
	}
	return td, cleanFunc, nil
}

type orgMemberInfo struct {
	orgId        string
	consensus    *testCertInfo
	admin        *testCertInfo
	client       *testCertInfo
	trustMember1 *testCertInfo
	trustMember2 *testCertInfo
}

type orgMember struct {
	orgId      string
	acProvider protocol.AccessControlProvider
	consensus  protocol.SigningMember
	admin      protocol.SigningMember
	client     protocol.SigningMember
}

var orgMemberInfoMap = map[string]*orgMemberInfo{
	testOrg1: {
		orgId:     testOrg1,
		consensus: testConsensusSignOrg1,
		admin:     testAdminSignOrg1,
		client:    testClientSignOrg1,
	},
	testOrg2: {
		orgId:     testOrg2,
		consensus: testConsensusSignOrg2,
		admin:     testAdminSignOrg2,
		client:    testClientSignOrg2,
	},
	testOrg3: {
		orgId:     testOrg3,
		consensus: testConsensusSignOrg3,
		admin:     testAdminSignOrg3,
		client:    testClientSignOrg3,
	},
	testOrg4: {
		orgId:     testOrg4,
		consensus: testConsensusSignOrg4,
		admin:     testAdminSignOrg4,
		client:    testClientSignOrg4,
	},
	testOrg5: {
		orgId:        testOrg5,
		consensus:    testConsensusSignOrg5,
		admin:        testAdminSignOrg5,
		client:       testClientSignOrg5,
		trustMember1: testTrustMemberAdmin1,
		trustMember2: testTrustMemberAdmin2,
	},
}

func initOrgMember(t *testing.T, info *orgMemberInfo) *orgMember {
	td, cleanFunc, err := createTempDirWithCleanFunc()
	require.Nil(t, err)
	defer cleanFunc()
	logger := logger2.GetLogger(logger2.MODULE_ACCESS)
	certProvider, err := newCertACProvider(testChainConfig, info.orgId, nil, logger)
	require.Nil(t, err)
	require.NotNil(t, certProvider)

	localPrivKeyFile := filepath.Join(td, info.orgId+".key")
	localCertFile := filepath.Join(td, info.orgId+".crt")

	err = ioutil.WriteFile(localPrivKeyFile, []byte(info.consensus.sk), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(localCertFile, []byte(info.consensus.cert), os.ModePerm)
	require.Nil(t, err)
	consensus, err := InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
	require.Nil(t, err)

	err = ioutil.WriteFile(localPrivKeyFile, []byte(info.admin.sk), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(localCertFile, []byte(info.admin.cert), os.ModePerm)
	require.Nil(t, err)
	admin, err := InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
	require.Nil(t, err)

	err = ioutil.WriteFile(localPrivKeyFile, []byte(info.client.sk), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(localCertFile, []byte(info.client.cert), os.ModePerm)
	require.Nil(t, err)
	client, err := InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
	require.Nil(t, err)

	return &orgMember{
		orgId:      info.orgId,
		acProvider: certProvider,
		consensus:  consensus,
		admin:      admin,
		client:     client,
	}
}

var mockAcLogger = logger.GetLogger(logger.MODULE_ACCESS)

func MockAccessControl() protocol.AccessControlProvider {
	certAc := &certACProvider{
		acService: &accessControlService{
			orgList:               &sync.Map{},
			orgNum:                0,
			resourceNamePolicyMap: &sync.Map{},
			hashType:              "",
			dataStore:             nil,
			memberCache:           concurrentlru.New(0),
			log:                   mockAcLogger,
			trustMembers:          nil,
		},
		certCache:  concurrentlru.New(0),
		crl:        sync.Map{},
		frozenList: sync.Map{},
		opts: bcx509.VerifyOptions{
			Intermediates: bcx509.NewCertPool(),
			Roots:         bcx509.NewCertPool(),
		},
		localOrg: nil,
		log:      mockAcLogger,
		hashType: "",
	}
	return certAc
}

func MockAccessControlWithHash(hashAlg string) protocol.AccessControlProvider {
	certAc := &certACProvider{
		acService: &accessControlService{
			orgList:               &sync.Map{},
			orgNum:                0,
			resourceNamePolicyMap: &sync.Map{},
			hashType:              hashAlg,
			dataStore:             nil,
			memberCache:           concurrentlru.New(0),
			log:                   mockAcLogger,
			trustMembers:          nil,
		},
		certCache:  concurrentlru.New(0),
		crl:        sync.Map{},
		frozenList: sync.Map{},
		opts: bcx509.VerifyOptions{
			Intermediates: bcx509.NewCertPool(),
			Roots:         bcx509.NewCertPool(),
		},
		localOrg: nil,
		log:      mockAcLogger,
		hashType: hashAlg,
	}
	return certAc
}

func MockSignWithMultipleNodes(msg []byte, signers []protocol.SigningMember, hashType string) (
	[]*commonPb.EndorsementEntry, error) {
	var ret []*commonPb.EndorsementEntry
	for _, signer := range signers {
		sig, err := signer.Sign(hashType, msg)
		if err != nil {
			return nil, err
		}
		signerSerial, err := signer.GetMember()
		if err != nil {
			return nil, err
		}
		ret = append(ret, &commonPb.EndorsementEntry{
			Signer:    signerSerial,
			Signature: sig,
		})
	}
	return ret, nil
}

func NewAccessControlWithChainConfig(localPrivKeyFile, localPrivKeyPwd, localCertFile string,
	chainConfig protocol.ChainConf, localOrgId string, store protocol.BlockchainStore, log protocol.Logger) (
	protocol.AccessControlProvider, error) {
	conf := chainConfig.ChainConfig()
	acp, err := newCertACProvider(conf, localOrgId, store, log)
	if err != nil {
		return nil, err
	}
	chainConfig.AddWatch(acp)
	chainConfig.AddVmWatch(acp)
	InitCertSigningMember(testChainConfig, localOrgId, localPrivKeyFile, localPrivKeyPwd, localCertFile)
	return acp, err
}

//func NewMemberFromCertPem(orgID ,certPEM string){
//}
