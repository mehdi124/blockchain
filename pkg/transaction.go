package pkg

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"

	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
)

type Transaction struct {
	ID []byte
	Vin [] TXInput
	Vout [] TXOutput
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

	wallets,err := NewWallets()
	if err != nil{
		log.Panic(err)
	}

	wallet := wallets.GetWallet(from)
	pubKeyHash := HashPubKey( wallet.PublicKey )

	acc , validOutputs := bc.FindSpendableOutputs(pubKeyHash,amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid , outs := range validOutputs {

		txID , err := hex.DecodeString(txid)
		if err != nil{
			log.Panic(err)
		}

		for _,out := range outs {
			input := TXInput{ Txid:txID , Vout:out , Signature:nil , PubKey:wallet.PublicKey }
			inputs := append(inputs,input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs,*NewTXOutput(amount,to))
	if acc > amount {
		outputs = append(outputs,*NewTXOutput(acc - amount,from))
	}

	tx := Transaction{nil,inputs,outputs}
	tx.ID = tx.Hash()

	bc.SignTransaction(&tx,wallet.PrivateKey)

	return &tx
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey,prevTXs map[string]Transaction){

	if tx.IsCoinbase(){
		return
	}

	for _,vin := range tx.Vin {

		if prevTXs[ hex.EncodeToString(vin.Txid) ].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID,vin := range txCopy.Vin{

		prevTX := prevTXs[ hex.EncodeToString(vin.Txid) ]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature

	}

}

func (tx *Transaction) Verify( prevTXs map[string]Transaction ) bool{

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID,vin := range tx.Vin {

		prevTX := prevTXs[ hex.EncodeToString(vin.Txid) ]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[ vin.Vout ].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen/2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}

		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}

		if ecdsa.Verify(&rawPubKey,txCopy.ID,&r,&s) == false {
			return false
		}

	}

	return true

}

func (tx *Transaction) Hash() []byte{

	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}



// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {

	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}


//A trimmed copy will be signed, not a full transaction
func (tx *Transaction) TrimmedCopy() Transaction {

	var inputs []TXInput
	var outputs []TXOutput

	for _,vin := range tx.Vin {
		inputs = append(inputs,TXInput{vin.Txid,vin.Vout,nil,nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID,inputs,outputs}
	return txCopy

}

