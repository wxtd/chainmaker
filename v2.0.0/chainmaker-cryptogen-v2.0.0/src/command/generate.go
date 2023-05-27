package command

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"chainmaker.org/chainmaker-cryptogen/config"
	"chainmaker.org/chainmaker/common/v2/cert"
	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"
	"chainmaker.org/chainmaker/common/v2/helper"
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"github.com/mr-tron/base58/base58"
	"github.com/spf13/cobra"
)

func GenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate key material",
		Long:  "Generate key material",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate()
		},
	}
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "crypto-config", "specify the output directory in which to place artifacts")
	return generateCmd
}

func generate() error {
	cryptoGenConfig := config.GetCryptoGenConfig()

	for _, item := range cryptoGenConfig.Item {
		for i := 0; i < int(item.Count); i++ {
			orgName := fmt.Sprintf("%s%d.%s", item.HostName, i+1, item.Domain)
			if item.Count == 1 {
				orgName = fmt.Sprintf("%s.%s", item.HostName, item.Domain)
			}
			keyType := crypto.AsymAlgoMap[strings.ToUpper(item.PKAlgo)]
			hashType := crypto.HashAlgoMap[strings.ToUpper(item.SKIHash)]

			caPath := filepath.Join(outputDir, orgName, "ca")
			caKeyPath := filepath.Join(caPath, "ca.key")
			caCertPath := filepath.Join(caPath, "ca.crt")
			userPath := filepath.Join(outputDir, orgName, "user")
			nodePath := filepath.Join(outputDir, orgName, "node")

			caCN := fmt.Sprintf("ca.%s", orgName)
			caSANS := append(item.CA.Specs.SANS, caCN)
			if err := generateCA(caPath,
				item.CA.Location.Country, item.CA.Location.Locality, item.CA.Location.Province, "root-cert", orgName, caCN,
				item.CA.Specs.ExpireYear, caSANS, keyType, hashType); err != nil {
				return err
			}

			for _, user := range item.User {
				for j := 0; j < int(user.Count); j++ {
					name := fmt.Sprintf("%s%d", user.Type, j+1)
					path := filepath.Join(userPath, name)
					if err := generateUser(path, name, caKeyPath, caCertPath,
						user.Location.Country, user.Location.Locality, user.Location.Province, orgName, user.Type,
						user.ExpireYear, keyType, hashType); err != nil {
						return err
					}
				}
			}

			for _, node := range item.Node {
				for j := 0; j < int(node.Count); j++ {
					name := fmt.Sprintf("%s%d", node.Type, j+1)
					path := filepath.Join(nodePath, name)
					if err := generateNode(path, name, caKeyPath, caCertPath,
						node.Location.Country, node.Location.Locality, node.Location.Province, orgName, node.Type,
						node.Specs.ExpireYear, node.Specs.SANS, keyType, hashType); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func generateCA(path, c, l, p, ou, org, cn string, expireYear int32, sans []string, keyType crypto.KeyType, hashType crypto.HashType) error {
	privKey, err := cert.CreatePrivKey(keyType, path, "ca.key")
	if err != nil {
		return err
	}

	return cert.CreateCACertificate(&cert.CACertificateConfig{PrivKey: privKey, HashType: hashType, CertPath: path, CertFileName: "ca.crt", Country: c, Locality: l, Province: p, OrganizationalUnit: ou, Organization: org, CommonName: cn, ExpireYear: expireYear, Sans: sans})
}

func generateUser(path, name, caKeyPath, caCertPath string,
	c, l, p, org, ou string, expireYear int32,
	keyType crypto.KeyType, hashType crypto.HashType) error {
	signName := fmt.Sprintf("%s.sign", name)
	signCN := fmt.Sprintf("%s.%s", signName, org)
	tlsName := fmt.Sprintf("%s.tls", name)
	tlsCN := fmt.Sprintf("%s.%s", tlsName, org)

	if err := generatePairs(path, caKeyPath, caCertPath,
		c, l, p, org, ou, signCN,
		signName, expireYear, []string{}, "", keyType, hashType); err != nil {
		return err
	}
	if err := generatePairs(path, caKeyPath, caCertPath,
		c, l, p, org, ou, tlsCN,
		tlsName, expireYear, []string{}, "", keyType, hashType); err != nil {
		return err
	}

	// calc client address for DPoS using client's sign cert
	if strings.HasPrefix(name, "client") {
		userCertBytes, err := ioutil.ReadFile(filepath.Join(path, fmt.Sprintf("%s.crt", signName)))
		if err != nil {
			return err
		}
		cert, err := parseCert(userCertBytes)
		if err != nil {
			return err
		}
		pubKey, err := cert.PublicKey.Bytes()
		if err != nil {
			return err
		}

		var (
			addr   string
			hashBz []byte
		)
		if cert.SignatureAlgorithm == bcx509.SM3WithSM2 {
			hashBz, err = hash.GetByStrType(crypto.CRYPTO_ALGO_SM3, pubKey)
		} else {
			hashBz, err = hash.GetByStrType(crypto.CRYPTO_ALGO_SHA256, pubKey)
		}
		if err != nil {
			return err
		}
		addr = base58.Encode(hashBz[:])
		userAddrFileName := filepath.Join(path, fmt.Sprintf("%s.addr", name))
		err = ioutil.WriteFile(userAddrFileName, []byte(addr), 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateNode(path, name, caKeyPath, caCertPath string,
	c, l, p, org, ou string, expireYear int32, sans []string,
	keyType crypto.KeyType, hashType crypto.HashType) error {
	signName := fmt.Sprintf("%s.sign", name)
	signCN := fmt.Sprintf("%s.%s", signName, org)
	tlsName := fmt.Sprintf("%s.tls", name)
	tlsCN := fmt.Sprintf("%s.%s", tlsName, org)
	tlsSANS := append(sans, tlsCN)

	id := uuid.GetUUID()
	if err := generatePairs(path, caKeyPath, caCertPath,
		c, l, p, org, ou, signCN,
		signName, expireYear, []string{}, id, keyType, hashType); err != nil {
		return err
	}
	if err := generatePairs(path, caKeyPath, caCertPath,
		c, l, p, org, ou, tlsCN,
		tlsName, expireYear, tlsSANS, id, keyType, hashType); err != nil {
		return err
	}

	certBytes, err := ioutil.ReadFile(filepath.Join(path, fmt.Sprintf("%s.crt", tlsName)))
	if err != nil {
		return err
	}
	nodeUid, err := helper.GetLibp2pPeerIdFromCert(certBytes)
	if err != nil {
		return err
	}
	nodeIdFileName := filepath.Join(path, fmt.Sprintf("%s.nodeid", name))
	return ioutil.WriteFile(nodeIdFileName, []byte(nodeUid), 0600)
}

func generatePairs(path, caKeyPath, caCertPath, c, l, p, org, ou, cn, name string, expireYear int32, sans []string, uuid string, keyType crypto.KeyType, hashType crypto.HashType) error {
	keyName := fmt.Sprintf("%s.key", name)
	csrName := fmt.Sprintf("%s.csr", name)
	certName := fmt.Sprintf("%s.crt", name)
	csrPath := filepath.Join(path, csrName)
	defer os.Remove(csrPath)

	privKey, err := cert.CreatePrivKey(keyType, path, keyName)
	if err != nil {
		return err
	}
	if err := cert.CreateCSR(&cert.CSRConfig{PrivKey: privKey, CsrPath: path, CsrFileName: csrName, Country: c, Locality: l, Province: p, OrganizationalUnit: ou, Organization: org, CommonName: cn}); err != nil {
		return err
	}
	if err := cert.IssueCertificate(&cert.IssueCertificateConfig{HashType: hashType, IssuerPrivKeyFilePath: caKeyPath, IssuerCertFilePath: caCertPath, CsrFilePath: csrPath, CertPath: path, CertFileName: certName, ExpireYear: expireYear, Sans: sans, Uuid: uuid}); err != nil {
		return err
	}

	return nil
}

// parseCert convert bytearray to certificate
func parseCert(crtPEM []byte) (*bcx509.Certificate, error) {
	certBlock, _ := pem.Decode(crtPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("decode pem failed, invalid certificate")
	}

	cert, err := bcx509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509 parse cert failed, %s", err)
	}

	return cert, nil
}
