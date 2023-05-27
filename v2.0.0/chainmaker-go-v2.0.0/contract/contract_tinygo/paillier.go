package main

const (
	PaillierOpTypeAddCiphertext = "AddCiphertext"
	PaillierOpTypeAddPlaintext  = "AddPlaintext"
	PaillierOpTypeSubCiphertext = "SubCiphertext"
	PaillierOpTypeSubPlaintext  = "SubPlaintext"
	PaillierOpTypeNumMul        = "NumMul"
)

type PaillierContext interface {
	AddCiphertext(pubKey []byte, ct1 []byte, ct2 []byte) ([]byte, ResultCode)
	AddPlaintext(pubKey, ct []byte, pt string) ([]byte, ResultCode)
	SubCiphertext(pubKey, ct1, ct2 []byte) ([]byte, ResultCode)
	SubPlaintext(pubKey, ct []byte, pt string) ([]byte, ResultCode)
	NumMul(pubKey, ct []byte, pt string) ([]byte, ResultCode)
}

type PaillierContextImpl struct{}

func NewPaillierContext() PaillierContext {
	return &PaillierContextImpl{}
}

func (p *PaillierContextImpl) AddCiphertext(pubKey, ct1, ct2 []byte) ([]byte, ResultCode) {
	return getPaillierResultBytes(pubKey, ct1, ct2, PaillierOpTypeAddCiphertext)
}

func (p *PaillierContextImpl) AddPlaintext(pubKey, ct []byte, pt string) ([]byte, ResultCode) {
	return getPaillierResultBytes(pubKey, ct, []byte(pt), PaillierOpTypeAddPlaintext)
}

func (p *PaillierContextImpl) SubCiphertext(pubKey, ct1, ct2 []byte) ([]byte, ResultCode) {
	return getPaillierResultBytes(pubKey, ct1, ct2, PaillierOpTypeSubCiphertext)
}

func (p *PaillierContextImpl) SubPlaintext(pubKey, ct []byte, pt string) ([]byte, ResultCode) {
	return getPaillierResultBytes(pubKey, ct, []byte(pt), PaillierOpTypeSubPlaintext)
}

func (p *PaillierContextImpl) NumMul(pubKey, ct []byte, pt string) ([]byte, ResultCode) {
	return getPaillierResultBytes(pubKey, ct, []byte(pt), PaillierOpTypeNumMul)
}

func getPaillierResultBytes(pubKey, operandOne, operandTwo []byte, opType string) ([]byte, ResultCode) {
	ec := NewEasyCodec()
	ec.AddString("opType", opType)
	ec.AddBytes("operandOne", operandOne)
	ec.AddBytes("operandTwo", operandTwo)
	ec.AddBytes("pubKey", pubKey)
	return GetBytesFromChain(ec, ContractMethodGetPaillierOperationResultLen, ContractMethodGetPaillierOperationResult)
}
