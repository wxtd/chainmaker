package main

const (
	BulletproofsOpTypePedersenAddNum        = "PedersenAddNum"
	BulletproofsOpTypePedersenAddCommitment = "PedersenAddCommitment"
	BulletproofsOpTypePedersenSubNum        = "PedersenSubNum"
	BulletproofsOpTypePedersenSubCommitment = "PedersenSubCommitment"
	BulletproofsOpTypePedersenMulNum        = "PedersenMulNum"
	BulletproofsVerify                      = "BulletproofsVerify"
)

// BulletproofsContext is the interface that wrap the bulletproofs method
type BulletproofsContext interface {
	// PedersenAddNum Compute a commitment to x + y from a commitment to x without revealing the value x, where y is a scalar
	// commitment: C = xB + rB'
	// value: the value y
	// return1: the new commitment to x + y: C' = (x + y)B + rB'
	PedersenAddNum(commitment []byte, num string) ([]byte, ResultCode)

	// PedersenAddCommitment Compute a commitment to x + y from commitments to x and y, without revealing the value x and y
	// commitment1: commitment to x: Cx = xB + rB'
	// commitment2: commitment to y: Cy = yB + sB'
	// return: commitment to x + y: C = (x + y)B + (r + s)B'
	PedersenAddCommitment(commitment1, commitment2 []byte) ([]byte, ResultCode)

	// PedersenSubNum Compute a commitment to x - y from a commitment to x without revealing the value x, where y is a scalar
	// commitment: C = xB + rB'
	// value: the value y
	// return1: the new commitment to x - y: C' = (x - y)B + rB'
	PedersenSubNum(commitment []byte, num string) ([]byte, ResultCode)

	// PedersenSubCommitment Compute a commitment to x - y from commitments to x and y, without revealing the value x and y
	// commitment1: commitment to x: Cx = xB + rB'
	// commitment2: commitment to y: Cy = yB + sB'
	// return: commitment to x - y: C = (x - y)B + (r - s)B'
	PedersenSubCommitment(commitment1, commitment2 []byte) ([]byte, ResultCode)

	// PedersenMulNum Compute a commitment to x * y from a commitment to x and an integer y, without revealing the value x and y
	// commitment1: commitment to x: Cx = xB + rB'
	// value: integer value y
	// return: commitment to x * y: C = (x * y)B + (r * y)B'
	PedersenMulNum(commitment []byte, num string) ([]byte, ResultCode)

	// Verify Verify the validity of a proof
	// proof: the zero-knowledge proof proving the number committed in commitment is in the range [0, 2^64)
	// commitment: commitment bindingly hiding the number x
	// return: true on valid proof, false otherwise
	Verify(proof, commitment []byte) ([]byte, ResultCode)
}

type BulletproofsContextImpl struct{}

func NewBulletproofsContext() BulletproofsContext {
	return &BulletproofsContextImpl{}
}

func (*BulletproofsContextImpl) PedersenAddNum(commitment []byte, num string) ([]byte, ResultCode) {
	return getBulletproofsResultBytes(commitment, []byte(num), BulletproofsOpTypePedersenAddNum)
}

func (*BulletproofsContextImpl) PedersenAddCommitment(commitment1, commitment2 []byte) ([]byte, ResultCode) {
	return getBulletproofsResultBytes(commitment1, commitment2, BulletproofsOpTypePedersenAddCommitment)
}

func (*BulletproofsContextImpl) PedersenSubNum(commitment []byte, num string) ([]byte, ResultCode) {
	return getBulletproofsResultBytes(commitment, []byte(num), BulletproofsOpTypePedersenSubNum)
}

func (*BulletproofsContextImpl) PedersenSubCommitment(commitment1, commitment2 []byte) ([]byte, ResultCode) {
	return getBulletproofsResultBytes(commitment1, commitment2, BulletproofsOpTypePedersenSubCommitment)
}

func (*BulletproofsContextImpl) PedersenMulNum(commitment []byte, num string) ([]byte, ResultCode) {
	return getBulletproofsResultBytes(commitment, []byte(num), BulletproofsOpTypePedersenMulNum)
}

func (*BulletproofsContextImpl) Verify(proof, commitment []byte) ([]byte, ResultCode) {
	return getBulletproofsResultBytes(proof, commitment, BulletproofsVerify)
}

func getBulletproofsResultBytes(param1, param2 []byte, bulletproofsFuncName string) ([]byte, ResultCode) {
	ec := NewEasyCodec()
	ec.AddBytes("param1", param1)
	ec.AddBytes("param2", param2)
	ec.AddString("bulletproofsFuncName", bulletproofsFuncName)
	return GetBytesFromChain(ec, ContractMethodGetBulletproofsResultLen, ContractMethodGetBulletproofsResult)
}
