package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"bytes"
	"encoding/gob"
)

type Transaction struct {
	ID []byte
	Vin [] TXInput
	Vout [] TXOutput
}


// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

// TXOutput represents a transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey string
}


const subsidy = 10

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to,data string) *Transaction {

	if data == ""{
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{ []byte{},-1,data }
	txout := TXOutput{subsidy,to}
	tx := Transaction{nil,[]TXInput{txin},[]TXOutput{txout}}
	tx.SetID()
	return &tx
}


// Hash returns the hash of the Transaction
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}


func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}


func NewUTXOTransaction(from,to string,amount int,bc *Blockchain) *Transaction {

	var inputs []TXInput
	var outputs []TXOutput

	acc,validOutputs := bc.FindSpendableOutputs(from,amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid,outs := range validOutputs {

		txID , _ := hex.DecodeString(txid)

		for _, out := range outs {
			input := TXInput{Txid:txID,Vout:out,ScriptSig:from}
			inputs = append(inputs,input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs,TXOutput{amount,to})
	if acc > amount {
		//back extra balance to from wallet (fucking amazing)
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil,inputs,outputs}
	tx.SetID()

	return &tx

}


