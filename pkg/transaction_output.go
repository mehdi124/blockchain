package pkg

import "bytes"

// TXOutput represents a transaction output
type TXOutput struct {
	Value        int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte){

	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(publicHash []byte) bool{
	return bytes.Compare(out.PubKeyHash,publicHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int,address string) *TXOutput {

	txo := &TXOutput{value,nil}
	txo.Lock( []byte(address) )
	return txo
}




