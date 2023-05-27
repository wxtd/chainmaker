PLATFORM=$(shell uname -m)
DATETIME=$(shell date "+%Y%m%d%H%M%S")
VERSION=v0.1.0

chainmaker-cryptogen:
	@cd src && GOPATH=${GOPATH} go build -o ../bin/chainmaker-cryptogen
	@mkdir -p ./release
	@rm -rf ./tmp/chainmaker-cryptogen/
	@mkdir -p ./tmp/chainmaker-cryptogen/
	@mkdir ./tmp/chainmaker-cryptogen/bin
	@mkdir ./tmp/chainmaker-cryptogen/config
	@cp bin/chainmaker-cryptogen ./tmp/chainmaker-cryptogen/bin
	@cp -r config/crypto_config_template.yml ./tmp/chainmaker-cryptogen/config
	@cd ./tmp;tar -zcvf chainmaker-cryptogen-$(VERSION).$(DATETIME).$(PLATFORM).tar.gz chainmaker-cryptogen; mv chainmaker-cryptogen-$(VERSION).$(DATETIME).$(PLATFORM).tar.gz ../release
	@rm -rf ./tmp/

clean:
	@rm -rf ./bin
